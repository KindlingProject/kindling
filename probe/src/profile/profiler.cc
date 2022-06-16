#include <cmath>
#include "profile/profiler.h"
#include "profile/flame_graph.h"
#include <cstddef>
#include <iostream>
#include <string.h>
#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>

using namespace std;

struct ProfileCtx {
    FlameGraph *flame_graph;
} profile_ctx;

void do_profile(struct sample_type_data *sample_data) {
    if (sample_data->tid_entry.pid == 0) {
        return;
    }
    profile_ctx.flame_graph->RecordSampleData(sample_data);
}

void do_collect() {
//    char str[50];
//    time_t now = time(NULL);
//    strftime(str, 50, "%x %X", localtime(&now));
//    cout << "===== " << str << " =====" << endl;
    
    profile_ctx.flame_graph->CollectData();
}

Profiler::Profiler(int cache_keep_time, int perf_period_ms) {
    profile_ctx.flame_graph = new FlameGraph(cache_keep_time, perf_period_ms);

    perf_data_ = (struct perfData *)malloc(sizeof(struct perfData) * 1);
    perf_data_->running = 0;
    perf_data_->sampleMs = perf_period_ms;
    perf_data_->sample = do_profile;
    perf_data_->collectMs = 1000;
    perf_data_->collect = do_collect;
}

Profiler::~Profiler() {
    free(perf_data_);

    delete profile_ctx.flame_graph;
    profile_ctx.flame_graph = nullptr;
}

void Profiler::Start() {
    perf(perf_data_);
}

void Profiler::Stop() {
    perf_data_->running = 0;
}

void Profiler::EnableAutoGet() {
    profile_ctx.flame_graph->EnableAutoGet();
}

void Profiler::EnableFlameFile() {
    profile_ctx.flame_graph->EnableFlameFile();
}

void Profiler::SetMaxDepth(int max_depth) {
    profile_ctx.flame_graph->SetMaxDepth(max_depth);
}

void Profiler::SetFilterThreshold(int filter_threshold) {
    profile_ctx.flame_graph->SetFilterThreshold(filter_threshold);
}

string Profiler::GetOnCpuData(__u32 tid, vector<pair<uint64_t, uint64_t>> &periods) {
    return profile_ctx.flame_graph->GetOnCpuData(tid, periods);
}