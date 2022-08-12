#include "profile/flame_graph.h"
#include "profile/bcc/symbol.h"
#include <string.h>

struct FlameGraphCtx{
    string seperator_frame_ = "-";
    string seperator_next_ = "#";
    string seperator_flame_ = "|";

    int max_depth_ = 1;
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

static void setProfileData(void* object, void* value) {
    ProfileData *profileData = (ProfileData*) object;
    ProfileData *newProfileData = (ProfileData*) value;

    profileData->tid_ = newProfileData->tid_;
    profileData->depth_ = newProfileData->depth_;
    profileData->finish_ = newProfileData->finish_;
    profileData->stack_ = newProfileData->stack_;
}

bool FlameSymbolDatas::addPerfSymbol(int depth, __u64 ip, int pid, bool user) {
    if (depth >= max_depth_) {
        return false;
    }
    FlameSymbolData* data = symbol_datas_.at(depth);
    data->symbol_ = flame_graph_ctx.symbol_table_->GetSymbol(ip, pid);
    data->type_ = user ? TYPE_USER : TYPE_KERNEL;
    symbol_to_ = depth;
    return true;
}

void FlameSymbolDatas::addProfileSymbol(int depth, string name) {
    if (depth >= max_depth_) {
        return;
    }
    if (depth == 0) {
        symbol_from_ = 0;
    } else if (depth != symbol_to_ + 1) {
        // FIX Skip First Data.
        symbol_from_ = depth;
    }
    FlameSymbolData* data = symbol_datas_.at(depth);
    data->symbol_ = name;
    data->type_ = TYPE_JVM;
    symbol_to_ = depth;
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

bool ProfileData::collectStacks(FlameSymbolDatas *symbolDatas) {
    int index = 0, depth = depth_;
    for (int i = 0; i < stack_.length(); i++) {
        if (stack_[i] == '!') {
            string symbol = stack_.substr(index, i - index);
            symbolDatas->addProfileSymbol(depth++, symbol);
            index = i + 1;
        }
    }
    return finish_;
}

void AggregateData::Aggregate() {
    // Sort by Desc.
    Node* f = root();
    for (int i = symbol_datas_->symbol_to_; i >= symbol_datas_->symbol_from_; i--) {
        f = f->addChild(symbol_datas_->symbol_datas_[i]->symbol_, symbol_datas_->symbol_datas_[i]->type_);
    }
    f->addLeaf();
}

void AggregateData::DumpFrameDatas(string& frameDatas, bool file) {
    if (root_.total_ < 1) {
        return;
    }
    dumpFrameData(frameDatas, "all", root_, 0, 0, false, file);
}

void AggregateData::dumpFrameData(string& frameDatas, const string& name, const Node& node, int depth, __u64 x, bool seperator, bool file) {
    Node node_copy = node;
    string name_copy = name;
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
    frameDatas.append(std::to_string(node_copy.getColor(name_copy)));
    frameDatas.append(flame_graph_ctx.seperator_frame_);
    if (file) {
        frameDatas.append("'");
        frameDatas.append(name_copy);
        frameDatas.append("')");
    } else {
        frameDatas.append(std::to_string(getFuncId(name_copy)));
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

static void aggSampleData(void* object, void* value) {
    AggregateData* agg_data = (AggregateData*) object;
    SampleData *sample_data = (SampleData*)value;
    if (sample_data->pid_ == 0 || (agg_data->tid_ > 0 && sample_data->tid_ != agg_data->tid_)) {
        return;
    }
    sample_data->collectStacks(agg_data->symbol_datas_);
    agg_data->Aggregate();
}

static void aggProfileData(void* object, void* value) {
    AggregateData* agg_data = (AggregateData*) object;
    ProfileData *profile_data = (ProfileData*)value;
    if (profile_data->tid_ == 0 || (profile_data->tid_ > 0 && profile_data->tid_ != profile_data->tid_)) {
        return;
    }

    bool finish = profile_data->collectStacks(agg_data->symbol_datas_);
    if (finish) {
        agg_data->Aggregate();
    }
}

FlameGraph::FlameGraph(int size, int cache_keep_ms, int perf_period_ms) {
    perf_period_ns_ = perf_period_ms * 1000000;
    sample_datas_ = new BucketRingBuffers<SampleData>(size, perf_period_ns_);
    perf_threshold_ns_ = perf_period_ns_;
    cache_keep_time_ = cache_keep_ms / perf_period_ms;
}

FlameGraph::~FlameGraph() {
    delete sample_datas_;
    delete flame_graph_ctx.symbol_table_;
    flame_graph_ctx.symbol_table_ = NULL;
}

void FlameGraph::SetMaxDepth(int max_depth) {
    if (max_depth > 1) {
        flame_graph_ctx.max_depth_ = max_depth;
    }
}

void FlameGraph::RecordSampleData(struct sample_type_data *sample_data) {
    if (sample_data->callchain.nr > 256) {
        //fprintf(stdout, "[Ignore Sample Data] Pid: %d, Tid: %d, Nr: %lld\n",sample_data->tid_entry.pid, sample_data->tid_entry.tid, sample_data->callchain.nr);
        return;
    }
    last_sample_time_ = sample_datas_->add(sample_data->time, sample_data, setSampleData);
}

void FlameGraph::RecordProfileData(uint64_t time, __u32 pid, __u32 tid, int depth, bool finish, string stack) {
    if (profile_datas_.count(pid) == 0) {
        profile_datas_[pid] = new BucketRingBuffers<ProfileData>(2000, perf_period_ns_);
    }
    ProfileData *data = new ProfileData(tid, depth, finish, stack);
    profile_datas_[pid]->addAndExpire(time, cache_keep_time_, data, setProfileData);
}

void FlameGraph::CollectData() {
    // Expire Perf Datas.
    sample_datas_->expire(last_sample_time_ - cache_keep_time_);
}

string FlameGraph::GetOnCpuData(__u32 pid, __u32 tid, vector<std::pair<uint64_t, uint64_t>> &periods) {
    string result = "";
    AggregateData *aggregateData = new AggregateData(tid, flame_graph_ctx.max_depth_);
    BucketRingBuffers<ProfileData> *profileData = NULL;
    if (profile_datas_.count(pid) > 0) {
        profileData = profile_datas_[pid];
    }
    __u64 start_time = 0, end_time = 0;
    __u64 size = periods.size();
    for (__u64 i = 0; i < size; i++) {
        if (periods[i].second - periods[i].first >= perf_threshold_ns_) {
            start_time = periods[i].first / perf_period_ns_; // ns->ms
            end_time = periods[i].second / perf_period_ns_; // ns->ms

            //fprintf(stdout, ">> Collect: %lld -> %lld, Duration: %lld, Exist Data %ld -> %ld\n", start_time, end_time, end_time-start_time, sample_datas_->getFrom(), sample_datas_->getTo());
            if (profileData != NULL) {
                profileData->collect(start_time, end_time, aggregateData, aggProfileData);
            } else {
                sample_datas_->collect(start_time, end_time, aggregateData, aggSampleData);
            }
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
    // fprintf(stdout, ">> FlameData: %s\n", result.c_str());
    return result;
}