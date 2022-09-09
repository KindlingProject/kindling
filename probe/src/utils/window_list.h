#pragma once

#include <vector>
#include <unordered_map>
#include <asm/types.h>

typedef void(*callback) (void*, void*);
typedef void(*setData) (void*, void*);
typedef long(*getTime) (void*);

template <typename Data>
class TimeBucket {
    private:
        long m_first;
        long m_last;
        std::vector<Data*> m_datas;
    public:
        TimeBucket() : m_first(0), m_last(0) {
            m_datas = {};
        }
        ~TimeBucket() {
            if (m_first > 0) {
                for (auto it = m_datas.begin(); it != m_datas.end();it++) {
                    delete *it;
                }
                m_datas.clear();
            }
        }

        void add(long ts, void* value) {
            if (m_first == 0) {
                m_first = ts;
            }
            m_last = ts;
            m_datas.push_back((Data*) value);
        }

        bool isEmpty() {
            return m_first == 0;
        }

        bool hasData(long fromTs, long toTs) {
            return m_last >= fromTs && m_first <= toTs;
        }

        void collect(void* obj, callback callFn, getTime getTimeFn, long fromTs, long toTs) {
            if (m_first > 0) {
                long time = 0;
                for (int i = 0; i < m_datas.size(); i++) {
                    time = (*getTimeFn)(m_datas[i]);
                    if (time >= fromTs && time < toTs) {
                        (*callFn)(obj, m_datas[i]);
                    }
                }
            }
        }

        void clear() {
            if (m_first > 0) {
                m_first = 0;
                m_last = 0;
                for (auto it = m_datas.begin(); it != m_datas.end();it++) {
                    delete *it;
                }
                m_datas.clear();
            }
        }
};

template <typename Data>
class SlidingWindow {
    private:
        // Check ExpireTimes.
        int m_times = 0;
        // Write Index.
        int m_write = 0;
        // N + 1 records.
        TimeBucket<Data>* m_buckets;

        int m_size;
    public:
        SlidingWindow(int size) {
            m_size = size + 1;
            m_buckets = new TimeBucket<Data>[m_size];
        }
        ~SlidingWindow() {
            delete []m_buckets;
            m_buckets = NULL;
        }

        void add(long ts, void* value) {
            m_buckets[m_write].add(ts, value);
        }

        int getExpireTimes() {
            if (m_buckets[m_write].isEmpty()) {
                // Reuse Bucket which store bucket not periodic. Eg. 1, 5, 6, 8, 10.
                m_times++;
                // fprintf(stdout, "[Ignore Expire] Index: %d\n", m_write);
            } else {
                int next = (m_write + 1) % m_size;
                // fprintf(stdout, "[Clear Data] Index: %d, ", next);
                m_buckets[next].clear();
                m_write = next;
                m_times = 0;
            }
            return m_times;
        }

        void collect(long fromTs, long toTs, void* obj, callback callFn, getTime getTimeFn) {
            // Check Prew Data as expire operation make writeIndex to next poistion.
            int prew = (m_write + m_size - 1) % m_size;
            if (m_buckets[prew].isEmpty() && m_buckets[m_write].isEmpty()) {
                return;
            }

            int to = m_write;
            int from = (m_write + 2) % m_size;
            if (from < to) {
                for (int i = from; i <= to; i++) {
                    TimeBucket<Data> *bucket = &m_buckets[i];
                    if (bucket->isEmpty() == false && bucket->hasData(fromTs, toTs)) {
                        bucket->collect(obj, callFn, getTimeFn, fromTs, toTs);
                    }
                }
            } else {
                for (int i = from; i < m_size; i++) {
                    TimeBucket<Data> *bucket = &m_buckets[i];
                    if (bucket->isEmpty() == false && bucket->hasData(fromTs, toTs)) {
                        bucket->collect(obj, callFn, getTimeFn, fromTs, toTs);
                    }
                }
                for (int i = 0; i <= to; i++) {
                    TimeBucket<Data> *bucket = &m_buckets[i];
                    if (bucket->isEmpty() == false && bucket->hasData(fromTs, toTs)) {
                        bucket->collect(obj, callFn, getTimeFn, fromTs, toTs);
                    }
                }
            }
        }
};

template <typename Data>
class WindowList {
    private:
        std::unordered_map<int64_t, SlidingWindow<Data>*> m_windows;
        int m_size;
    public:
        WindowList(int size) {
            m_size = size;
        }

        ~WindowList() {
            for (auto it = m_windows.begin(); it != m_windows.end();) {
                delete it->second;
                m_windows.erase(it++);
            }
        }

        void add(__u32 tid, long ts, void* value) {
            auto it = m_windows.find(tid);
            SlidingWindow<Data>* window;
            if (it == m_windows.end()) {
                window = new SlidingWindow<Data>(m_size);
                m_windows[tid] = window;
            } else {
                window = it->second;
            }
            window->add(ts, value);
        }

        SlidingWindow<Data>* find(__u32 tid) {
            auto it = m_windows.find(tid);
            if (it == m_windows.end()) {
                return NULL;
            }

            return it->second;
        }

        void checkExpire(int seconds) {
            for (auto it = m_windows.begin(); it != m_windows.end();) {
                SlidingWindow<Data>* window = it->second;
                // Remove Unused Thread datas.
                if (window->getExpireTimes() > seconds) {
                    // fprintf(stdout, "[Expire SlideWindow] tid: %ld\n", it->first);
                    delete window;
                    m_windows.erase(it++);
                } else {
                    it++;
                }
            }
        }
};
