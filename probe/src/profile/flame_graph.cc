#include "profile/flame_graph.h"
#include "profile/bcc/symbol.h"
#include <string.h>

struct FlameGraphCtx{
    string seperator_stack_ = "#";
    string seperator_count_ = "@";
    string seperator_next_ = "!";

    int max_depth_ = 1;
    bool auto_get_ = false;
    BPFSymbolTable *symbol_table_ = new BPFSymbolTable();;
} flame_graph_ctx;

static void setSampleData(void* object, void* value) {
    SampleData* sampleData = (SampleData*) object;
    sample_type_data *sample_data = (sample_type_data*)value;

    sampleData->pid_ = sample_data->tid_entry.pid;
    sampleData->tid_ = sample_data->tid_entry.tid;
    sampleData->nr_ = sample_data->callchain.nr;
    memcpy(&sampleData->ips_[0], sample_data->callchain.ips, sampleData->nr_ * sizeof(sample_data->callchain.ips[0]));
}

SampleData::SampleData() {}

string SampleData::addThreadInfo(string call_stack) {
    if (flame_graph_ctx.auto_get_) {
        call_stack.insert(0, flame_graph_ctx.seperator_stack_);
        call_stack.insert(0, std::to_string(tid_));
        call_stack.insert(0, flame_graph_ctx.seperator_stack_);
        call_stack.insert(0, std::to_string(pid_));
    }
    return call_stack;
}

string SampleData::GetString() {
    string call_stack = "";
    bool kernel = false, user = false, seperator = false;
    int depth = 0, pid = 0;
    __u64 ip = 0;
    for (__u64 i = 0; i < nr_; i++) {
        ip = ips_[i];
        if (ip == PERF_CONTEXT_KERNEL) {
            kernel = true;
            user = false; 
            pid = -1;
            continue;
        } else if (ip == PERF_CONTEXT_USER) {
            kernel = false;
            user = true;
            pid = pid_;
            continue;
        }
        if (kernel || user) {
            if (seperator) {
                call_stack.insert(0, flame_graph_ctx.seperator_stack_);
            } else {
                seperator = true;
            }
            call_stack.insert(0, flame_graph_ctx.symbol_table_->GetSymbol(ip, pid));
            if (++depth >= flame_graph_ctx.max_depth_) {
                return addThreadInfo(call_stack);
            }
        }
    }

    return addThreadInfo(call_stack);
}

AggregateData::AggregateData(__u32 tid) {
    tid_ = tid;
}

AggregateData::~AggregateData() {
    agg_count_map_.clear();
}

void AggregateData::Aggregate(void* data) {
    SampleData *sample_data = (SampleData*)data;
    if (sample_data->pid_ == 0 || (tid_ > 0 && sample_data->tid_ != tid_)) {
        return;
    }

    string call_stack = sample_data->GetString();
    auto itr = agg_count_map_.find(call_stack);
    if (itr == agg_count_map_.end()) {
        agg_count_map_.insert(std::make_pair(call_stack, 1));
    } else {
        itr->second = itr->second + 1;
    }
}

void AggregateData::Reset() {
    agg_count_map_.clear();
}

string AggregateData::ToString() {
    if (agg_count_map_.empty()) {
        return "";
    }

    string result = "";
    bool seperator = false;
    map<string, int>::iterator itr;
    for (itr = agg_count_map_.begin(); itr != agg_count_map_.end(); itr++) {
        if (seperator) {
            result.append(flame_graph_ctx.seperator_next_);
        } else {
            seperator = true;
        }
        result.append(itr->first);
        result.append(flame_graph_ctx.seperator_count_);
        result.append(std::to_string(itr->second));
    }
    return result;
}

static void aggTidData(void* object, void* value) {
    AggregateData* pObject = (AggregateData*) object;
    pObject->Aggregate(value);
}

FlameGraph::FlameGraph(int cache_keep_time, int perf_period_ms) {
    perf_period_ns_ = perf_period_ms * 1000000;
    sample_datas_ = new BucketRingBuffers<SampleData>(2000, perf_period_ns_);
    perf_threshold_ns_ = perf_period_ns_;
    cache_keep_time_ = cache_keep_time / perf_period_ms;
}

FlameGraph::~FlameGraph() {
    delete sample_datas_;
    delete flame_graph_ctx.symbol_table_;
    flame_graph_ctx.symbol_table_ = NULL;
}

void FlameGraph::EnableAutoGet() {
    flame_graph_ctx.auto_get_ = true;
}

void FlameGraph::EnableFlameFile() {
    write_flame_graph_ = true;
    flame_graph_ctx.seperator_stack_ = ";";
    flame_graph_ctx.seperator_count_ = " ";
    flame_graph_ctx.seperator_next_ = "\n";

    resetLogFile();
}

void FlameGraph::SetMaxDepth(int max_depth) {
    if (max_depth > 1) {
        flame_graph_ctx.max_depth_ = max_depth;
    }
}

void FlameGraph::SetFilterThreshold(int filter_threshold) {
    perf_threshold_ns_ = perf_period_ns_ * filter_threshold;
}

void FlameGraph::RecordSampleData(struct sample_type_data *sample_data) {
    if (sample_data->callchain.nr > 256) {
        //fprintf(stdout, "[Ignore Sample Data] Pid: %d, Tid: %d, Nr: %lld\n",sample_data->tid_entry.pid, sample_data->tid_entry.tid, sample_data->callchain.nr);
        return;
    }
    last_sample_time_ = sample_datas_->add(sample_data->time, sample_data, setSampleData);
}

void FlameGraph::CollectData() {
    if (flame_graph_ctx.auto_get_) {
        AggregateData *aggregateData = new AggregateData(0);
        sample_datas_->collect(last_collect_time_, last_sample_time_, aggregateData, aggTidData);
        if (write_flame_graph_) {
            //fprintf(collect_file_, "%s\n", aggregateData->ToString().c_str());  // Write To File.
            fclose(collect_file_);
            resetLogFile();
        } else {
            aggregateData->ToString();
        }
    }

    //fprintf(stdout, "Before Exipre Size: %d\n", sample_datas_->size());
    // Expire BucketRingBuffers Datas.
    sample_datas_->expire(last_sample_time_ - cache_keep_time_);
    //fprintf(stdout, "After Exipre Size: %d\n", sample_datas_->size());
    last_collect_time_ = last_sample_time_;
}

string FlameGraph::GetOnCpuData(__u32 tid, vector<std::pair<uint64_t, uint64_t>> &periods) {
    string result = "";
    AggregateData *aggregateData = new AggregateData(tid);

    __u64 start_time = 0, end_time = 0;
    __u64 size = periods.size();
    for (__u64 i = 0; i < size; i++) {
        if (periods[i].second - periods[i].first >= perf_threshold_ns_) {
            start_time = periods[i].first / perf_period_ns_; // ns->ms
            end_time = periods[i].second / perf_period_ns_; // ns->ms

            //fprintf(stdout, ">> Collect: %lld -> %lld, Duration: %lld, Exist Data %ld -> %ld\n", start_time, end_time, end_time-start_time, sample_datas_->getFrom(), sample_datas_->getTo());
            sample_datas_->collect(start_time, end_time, aggregateData, aggTidData);
            result.append(aggregateData->ToString());
            aggregateData->Reset();
        }
        result.append("|");
    }
    if (result.size() == size) {
        // Set values like |||| to empty.
        return "";
    }
    return result;
}

void FlameGraph::resetLogFile() {
    char collectFileName[128];
    time_t nowtime = time(NULL);
    tm *now = localtime(&nowtime);
    snprintf(collectFileName, sizeof(collectFileName), "flamegraph_%d_%d_%d.txt", now->tm_hour, now->tm_min, now->tm_sec);
    collect_file_ = fopen(collectFileName, "w+");
}
