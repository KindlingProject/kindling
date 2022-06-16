#pragma once

#include <iostream>
#include <list>

typedef void(*callback) (void*, void*);
typedef void(*setData) (void*, void*);
typedef long(*getTime) (void*);

template <typename Data>
class RingBuffer {
    private:
        int m_from;
        int m_next;
        int m_count;
        Data* m_data;
    public:
        int m_size;

        RingBuffer(int size):m_from(0),m_next(0),m_count(0) {
            m_size = size;
            m_data = new Data[size];
        }
        ~RingBuffer() {
            delete []m_data;
            m_data = NULL;
        }

        int push(void* value, setData setFn) {
            if (m_count == m_size) {
                return -1;
            }
            
            setFn(&m_data[m_next], value);
            m_next = (m_next + 1) % m_size;
            m_count++;
            return (m_next > 0) ? m_next - 1 : m_size - 1;
        }

        bool isFull() {
            return m_count == m_size;
        }

        bool isEmpty() {
            return m_count == 0;
        }

        int getCount() {
            return m_count;
        }

        void reset(int from, int to) {
            m_from = (to + 1) % m_size;
            if (to >= from) {
                m_count -= (to - from + 1);
            } else {
                m_count -= (to + m_size - from + 1);
            }
        }

        void onData(void* obj, callback callFn, int fromIdx, int toIdx) {
            if (m_count > 0) {
                // Called by Bucket whose data is sorted.
                for (int i = fromIdx; i <= toIdx; i++) {
                    (*callFn)(obj, &m_data[i]);
                }
            }
        }

        void onData(void* obj, callback callFn, getTime getTimeFn, int fromIdx, int toIdx, long fromTs, long toTs) {
            if (m_count > 0) {
                long time = 0;
                for (int i = fromIdx; i <= toIdx; i++) {
                    time = (*getTimeFn)(&m_data[i]);
                    if (time >= fromTs && time < toTs) {
                        (*callFn)(obj, &m_data[i]);
                    }
                }
            }
        }
};

template <typename Data>
class Bucket {
    private:
        long m_ts;
        RingBuffer<Data> *m_ring;
        int m_from;
        int m_to;
    public:
        Bucket() {}
        ~Bucket() {
            m_ring = NULL;
        }

        void setValue(long ts, RingBuffer<Data> *ring, int from, int to) {
            m_ts = ts;
            m_ring = ring;
            m_from = from;
            m_to = to;
        }

        long getTs() {
            return m_ts;
        }

        RingBuffer<Data> *getRingBuffer() {
            return m_ring;
        }

        void clearRingBuffer() {
            m_ring = NULL;
        }

        void setTo(int to) {
            m_to = to;
        }

        int getFrom() {
            return m_from;
        }

        int getTo() {
            return m_to;
        }

        int size() {
            if (m_from <= m_to) {
                return m_to - m_from + 1;
            }
            return m_to + m_ring->m_size - m_from + 1;
        }
};

template <typename Data>
class BucketCache {
    std::list<Bucket<Data>*> m_buckets;
    public:
        BucketCache() {
            m_buckets = {};
        }
        ~BucketCache() {
            m_buckets.clear();
        }

        Bucket<Data>* borrowBucket(RingBuffer<Data> *ringBuffer, long ts, int index) {
            Bucket<Data>* bucket;
            if (m_buckets.empty()) {
                bucket = new Bucket<Data>();
            } else {
                bucket = m_buckets.back();
                m_buckets.pop_back();
            }
            bucket->setValue(ts, ringBuffer, index, index);
            return bucket;
        }

        void returnBucket(Bucket<Data>* bucket) {
            bucket->clearRingBuffer();
            m_buckets.push_back(bucket);
        }
};

template <typename Data>
class BucketRingBuffers {
    private:
        int m_size;
        long m_bucket_period;
        std::list<Bucket<Data>*> m_buckets;
        RingBuffer<Data> *m_big_ring;
        BucketCache<Data> *m_bucket_cache;

        Bucket<Data> *addBucket(RingBuffer<Data> *ringBuffer, long ts, void* value, setData setFn) {
            Bucket<Data> *bucket = m_bucket_cache->borrowBucket(ringBuffer, ts, ringBuffer->push(value, setFn));
            m_buckets.push_back(bucket);
            return bucket;
        }
    public:
        BucketRingBuffers(int size, long bucket_period):m_size(size), m_bucket_period(bucket_period) {
            m_buckets = {};
            m_big_ring = new RingBuffer<Data>(size);
            m_bucket_cache = new BucketCache<Data>();
        }
        ~BucketRingBuffers() {
            m_buckets.clear();
            delete m_big_ring;
            delete m_bucket_cache;
        }
        
        long add(long ts, void* value, setData setFn) {
            long bucketTs = ts / m_bucket_period;
            if (m_buckets.empty()) {
                addBucket(m_big_ring, bucketTs, value, setFn);
            } else {
                Bucket<Data> *bucket = m_buckets.back();
                RingBuffer<Data> *ringBuffer = bucket->getRingBuffer();
                if (ringBuffer->isFull()) {
                    // RingBuffer is Full, Add Bucket for new RingBuffer()
                    RingBuffer<Data> *newBuffer;
                    if (m_big_ring->isEmpty()) {
                        newBuffer = m_big_ring;
                    } else {
                        newBuffer = new RingBuffer<Data>(m_size / 2);
                    }
                    addBucket(newBuffer, bucketTs, value, setFn);
                } else {
                    int index = ringBuffer->push(value, setFn);
                    if (bucket->getTs() == bucketTs) {
                        bucket->setTo(index);
                    } else {
                        m_buckets.push_back(m_bucket_cache->borrowBucket(ringBuffer, bucketTs, index));
                    }
                }
            }
            return bucketTs;
        }

        void addAndExpire(long ts, long cacheTs, void* value, setData setFn) {
            long bucketTs = add(ts, value, setFn);
            expire(bucketTs - cacheTs);
        }

        void expire(long bucketTs) {
            for (auto it = m_buckets.begin(); it != m_buckets.end();) {
                Bucket<Data> *bucket = *it;
                if (bucket->getTs() <= bucketTs) {
                    m_buckets.erase(it++);
                    bucket->getRingBuffer()->reset(bucket->getFrom(), bucket->getTo());

                    if (bucket->getRingBuffer()->isEmpty()) {
                        if (bucket->getRingBuffer() != m_big_ring) {
                            delete bucket->getRingBuffer();
                        }
                    }
                    m_bucket_cache->returnBucket(bucket);
                } else {
                    it++;
                }
            }
        }

        void collect(long fromBucketTs, long toBucketTs, void* obj, callback callFn) {
            for (auto it = m_buckets.begin(); it != m_buckets.end();it++) {
                Bucket<Data> *bucket = *it;
                if ((fromBucketTs == toBucketTs && bucket->getTs() == fromBucketTs) || (bucket->getTs() > fromBucketTs && bucket->getTs() <= toBucketTs)) {
                    bucket->getRingBuffer()->onData(obj, callFn, bucket->getFrom(), bucket->getTo());
                }
            }
        }

        void collect(long from, long to, void* obj, callback callFn, getTime getTimeFn) {
            long fromBucketTs = from / m_bucket_period;
            long toBucketTs = to / m_bucket_period;
            for (auto it = m_buckets.begin(); it != m_buckets.end();it++) {
                Bucket<Data> *bucket = *it;
                if (bucket->getTs() >= fromBucketTs && bucket->getTs() <= toBucketTs) {
                    bucket->getRingBuffer()->onData(obj, callFn, getTimeFn, bucket->getFrom(), bucket->getTo(), from, to);
                }
            }
        }

        int size() {
            int size = 0;
            for (auto it = m_buckets.begin(); it != m_buckets.end();it++) {
                Bucket<Data> *bucket = *it;
                size += bucket->size();
            }
            return size;
        }

        long getFrom() {
            if (m_buckets.empty()) {
                return 0;
            }
            return m_buckets.front()->getTs();
        }

        long getTo() {
            if (m_buckets.empty()) {
                return 0;
            }
            return m_buckets.back()->getTs();
        }
};
