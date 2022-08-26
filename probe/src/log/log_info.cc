#include "sinsp.h"
#include "log_info.h"
#include "utils/ring_buffer.h"
#include <cstdio>
#include <cstring>

LogData::LogData() {}

LogData::~LogData() {
    string().swap(data_);
}

void LogData::setData(long ts, int size, __u32 tid, char* data) {
    ts_ = ts;
    tid_ = tid;
    data_ = data;
    // fprintf(stdout, "[Add Log] Time: %ld, Tid: %d, Data(%d): %s\n", ts, tid, size, data);
}

long LogData::getTs() {
    return ts_;
}

__u32 LogData::getTid() {
    return tid_;
}

string LogData::getData() {
    return data_;
}

static void setLog(void* object, void* evt) {
    LogData *logData = (LogData*) object;
    sinsp_evt *sEvt = (sinsp_evt*) evt;

    // Get Thread Id
    auto s_tinfo = sEvt->get_thread_info();
    auto pData = sEvt->get_param_value_raw("data");

    if (pData->m_len > 0) {
        char* log_info = new char[pData->m_len];
        memcpy(log_info, pData->m_val, pData->m_len);
        log_info[pData->m_len - 1] = '\0';
        logData->setData(sEvt->get_ts(), pData->m_len, s_tinfo->m_tid, log_info);
        delete []log_info;
    }
}

static long getLogTime(void* object) {
    LogData *logData = (LogData*) object;
    return logData->getTs();
}

LogDatas::LogDatas(__u32 tid) {
    tid_ = tid;
}

LogDatas::~LogDatas() {
    logs_.clear();
}

void LogDatas::CollectLogs(void* data) {
    LogData *log_data = (LogData*)data;
    if (log_data->getTid() != tid_) {
        return;
    }
    //fprintf(stdout, "Collect Log: %s\n", log_data->getData().c_str());
    // TODO log is split to 2 logs.
    logs_.push_back(log_data->getData());
}

void LogDatas::Reset() {
    logs_.clear();
}

string LogDatas::ToString() {
    if (logs_.empty()) {
        return "";
    }

    string result = "";
    bool seperator = false;
    for (auto itr = logs_.begin(); itr != logs_.end(); itr++) {
        if (seperator) {
            result.append("<br>");
        } else {
            seperator = true;
        }
        result.append(*itr);
    }
    return result;
}

static void collectTidData(void* object, void* value) {
    LogDatas* pObject = (LogDatas*) object;
    pObject->CollectLogs(value);
}

LogCache::LogCache(int size, int cache_ms) {
    long bucketTs = 10000000; // 10Ms
    cacheBucketTime = (1000000l * cache_ms) / bucketTs;
    logs_ = new BucketRingBuffers<LogData>(size, bucketTs);
}

LogCache::~LogCache() {
    delete logs_;
}

static bool isLogFile(sinsp_evt *sEvt) {
    auto s_fdinfo = sEvt->get_fd_info();
    if (!s_fdinfo) {
        return false;
    }
    if (s_fdinfo->m_type == SCAP_FD_FILE || s_fdinfo->m_type == SCAP_FD_FILE_V2) {
        // xxx.log xxx.log.yyyy-mm-dd-index
        return s_fdinfo->m_name.find(".log") != string::npos;
    }
    return false;
}

bool LogCache::addLog(void *evt) {
    sinsp_evt *sEvt = (sinsp_evt *) evt;
    // Check is write log.
    auto type = sEvt->get_type();
    if (type != PPME_SYSCALL_WRITE_X && type != PPME_SYSCALL_WRITEV_X
            && type != PPME_SYSCALL_PWRITE_X && type != PPME_SYSCALL_PWRITEV_X ) {
        return false;
    }
    if (isLogFile(sEvt) == false) {
        return false;
    }
    
    // Get Thread Id
    auto s_tinfo = sEvt->get_thread_info();
    if (!s_tinfo) {
        return false;
    }
    auto pres = sEvt->get_param_value_raw("res");
    if (!pres || *(int64_t *) pres->m_val <= 0) {
        return false;
    }
    count++;
    logs_->addAndExpire(sEvt->get_ts(), cacheBucketTime, evt, setLog);
    if (count % 10000 == 0) {
        fprintf(stdout, "Log Count: %d # %ld\n", logs_->size(), count);
    }
    return true;
}

string LogCache::getLogs(__u32 tid, vector<std::pair<uint64_t, uint64_t>> &periods) {
    __u64 size = periods.size();
    LogDatas *logDatas = new LogDatas(tid);
    // len@logs|len2@logs|0@|
    string result = "";
    for (__u64 i = 0; i < size; i++) {            
        logs_->collect(periods[i].first, periods[i].second, logDatas, collectTidData, getLogTime);
        string log = logDatas->ToString();
        if (log != "") {
            result.append(std::to_string(log.length()));
            result.append("@");
            result.append(log);
            result.append("|");
        } else {
            result.append("0@|");
        }
        logDatas->Reset();
    }
    delete logDatas;
    return result.length() == periods.size() * 3 ? "" : result;
}
