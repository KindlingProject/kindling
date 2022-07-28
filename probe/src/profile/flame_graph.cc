#include "profile/flame_graph.h"
#include "profile/bcc/symbol.h"
#include <string.h>

struct FlameGraphCtx{
    string seperator_frame_ = "-";
    string seperator_next_ = "#";
    string seperator_flame_ = "|";

    int max_depth_ = 1;
    bool auto_get_ = false;
    BPFSymbolTable *symbol_table_ = new BPFSymbolTable();
} flame_graph_ctx;

static void setSampleData(void* object, void* value) {
    SampleData* sampleData = (SampleData*) object;
    sample_type_data *sample_data = (sample_type_data*)value;

    sampleData->pid_ = sample_data->tid_entry.pid;
    sampleData->tid_ = sample_data->tid_entry.tid;
    sampleData->nr_ = sample_data->callchain.nr;
    memcpy(&sampleData->ips_[0], sample_data->callchain.ips, sampleData->nr_ * sizeof(sample_data->callchain.ips[0]));
}

bool FlameSymbolDatas::addPerfSymbol(int depth, __u64 ip, int pid, bool user) {
    if (depth >= max_depth_) {
        return false;
    }
    FlameSymbolData* data = symbol_datas_.at(depth);
    data->symbol_ = flame_graph_ctx.symbol_table_->GetSymbol(ip, pid);
    data->color_ = getPerfColor(user);
    symbol_size_ = depth;
    return true;
}

void SampleData::collectStacks(FlameSymbolDatas *symbolDatas) {
    bool kernel = false, user = false;
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
            if (symbolDatas->addPerfSymbol(depth++, ip, pid, user) == false) {
                return;
            }
        }
    }
}

void AggregateData::Aggregate(void* data) {
    SampleData *sample_data = (SampleData*)data;
    if (sample_data->pid_ == 0 || (tid_ > 0 && sample_data->tid_ != tid_)) {
        return;
    }

    Node* f = root();
    if (tid_ == 0) {
        f = f->addChild(std::to_string(sample_data->pid_), getPerfColor(false));
        f = f->addChild(std::to_string(sample_data->tid_), getPerfColor(false));
    }

    sample_data->collectStacks(symbol_datas_);
    // Sort by Desc.
    for (int i = symbol_datas_->symbol_size_ - 1; i >= 0; i--) {
        f = f->addChild(symbol_datas_->symbol_datas_[i]->symbol_, symbol_datas_->symbol_datas_[i]->color_);
    }
    f->addLeaf();
}

void AggregateData::DumpFrameDatas(string& frameDatas, bool file) {
    if (root_.total_ == 0) {
        return;
    }
    dumpFrameData(frameDatas, "all", root_, 0, 0, false, file);
}

void AggregateData::dumpFrameData(string& frameDatas, const string& name, const Node& node, int depth, __u64 x, bool seperator, bool file) {
    if (seperator) {
        frameDatas.append(flame_graph_ctx.seperator_next_);
    }
    if (file) {
        frameDatas.append("f(");
    }
    frameDatas.append(std::to_string(depth));
    frameDatas.append(flame_graph_ctx.seperator_frame_);
    frameDatas.append(std::to_string(x));
    frameDatas.append(flame_graph_ctx.seperator_frame_);
    frameDatas.append(std::to_string(node.total_));
    frameDatas.append(flame_graph_ctx.seperator_frame_);
    frameDatas.append(std::to_string(node.color_));
    frameDatas.append(flame_graph_ctx.seperator_frame_);
    if (file) {
        frameDatas.append("'");
        frameDatas.append(name);
        frameDatas.append("')");
    } else {
        frameDatas.append(std::to_string(getFuncId(name)));
    }

    x += node.self_;
    for (map<string, Node>::const_iterator itr = node.children_.begin(); itr != node.children_.end(); ++itr) {
        if (itr->second.total_ > 2) {
            dumpFrameData(frameDatas, itr->first, itr->second, depth + 1, x, true, file);
        }
        x += itr->second.total_;
    }
}

void AggregateData::DumpFuncNames(string& frameDatas) {
    if (func_names_.size() == 0) {
        return;
    }

    frameDatas.insert(0, flame_graph_ctx.seperator_flame_);
    for (size_t i = func_names_.size() - 1; i > 0; i--) {
        frameDatas.insert(0, func_names_[i]);
        frameDatas.insert(0, flame_graph_ctx.seperator_next_);
    }
    frameDatas.insert(0, func_names_[0]);
}


int AggregateData::getFuncId(const string& name) {
    int id = func_name_map_[name];
    if (id == 0) {
        id = func_name_map_[name] = func_name_map_.size();
        func_names_.push_back(name);
    }
    return id - 1;
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
    flame_graph_ctx.seperator_frame_ = ", ";
    flame_graph_ctx.seperator_next_ = "\n";
    flame_graph_ctx.seperator_flame_ = "\n\n";

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
        AggregateData *aggregateData = new AggregateData(0, flame_graph_ctx.max_depth_);
        sample_datas_->collect(last_collect_time_, last_sample_time_, aggregateData, aggTidData);
        string result = "";
        aggregateData->DumpFrameDatas(result, write_flame_graph_);
        if (write_flame_graph_ == false) {
            aggregateData->DumpFuncNames(result);
        }

        if (write_flame_graph_) {
            fprintf(collect_file_, "%s\n", result.c_str());  // Write To File.
            fclose(collect_file_);
            resetLogFile();
        } else {
            fprintf(stdout, "%s\n", result.c_str());
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
    AggregateData *aggregateData = new AggregateData(tid, flame_graph_ctx.max_depth_);

    __u64 start_time = 0, end_time = 0;
    __u64 size = periods.size();
    for (__u64 i = 0; i < size; i++) {
        if (periods[i].second - periods[i].first >= perf_threshold_ns_) {
            start_time = periods[i].first / perf_period_ns_; // ns->ms
            end_time = periods[i].second / perf_period_ns_; // ns->ms

            //fprintf(stdout, ">> Collect: %lld -> %lld, Duration: %lld, Exist Data %ld -> %ld\n", start_time, end_time, end_time-start_time, sample_datas_->getFrom(), sample_datas_->getTo());
            sample_datas_->collect(start_time, end_time, aggregateData, aggTidData);
            aggregateData->DumpFrameDatas(result, false);
            aggregateData->Reset();
        }
        result.append(flame_graph_ctx.seperator_flame_);
    }

    if (result.size() == size) {
        // Set values like |||| to empty.
        return "";
    }

    aggregateData->DumpFuncNames(result);
    return result;
}

void FlameGraph::resetLogFile() {
    char collectFileName[128];
    time_t nowtime = time(NULL);
    tm *now = localtime(&nowtime);
    snprintf(collectFileName, sizeof(collectFileName), "flamegraph_%d_%d_%d.txt", now->tm_hour, now->tm_min, now->tm_sec);
    collect_file_ = fopen(collectFileName, "w+");
}
