#include <stdio.h>
#include <iostream>

#include "ring_buffer.h"

using namespace std;

struct Value {
    int digits;
    int tens;

    Value(int tens, int digits):digits(digits),tens(tens) {
    }
};

class Int {
    public:
        long m_ts;
        int m_value;

        Int():m_value(0) {}
        Int(long ts, int value):m_ts(ts), m_value(value) {}
        int getValue() {
            return m_value;
        }
        long getTs() {
            return m_ts;
        }
};

class Print {
    public:
        Print() {}
        void print(void *intValue) {
            Int intData = * (Int*)intValue;
            fprintf(stdout, "Data: %d at %ld\n", intData.getValue(), intData.getTs());
        }
};

static void printData(void* object, void* value) {
    Print* pObject = (Print*) object;
    pObject->print(value);
}

static void setInt(void* object, void* value) {
    Int* intVal = (Int*) object;
    Value *val = (Value*)value;
    intVal->m_ts = val->tens * 10;
    intVal->m_value = val->tens * 10 + val->digits;
}

static long getDataTime(void* value) {
    Int* intVal = (Int*) value;
    return intVal->getValue() * 1000000;
}

void testBucketRings() {
    BucketRingBuffers<Int> *rings = new BucketRingBuffers<Int>(20, 1);

    /**
     * Ring-0 0  [ 0,  9] 10 [10, 19]
     * Ring-1 20 [20, 29] 
     * Ring-2 30 [30, 39]
     */
    for (int i = 0; i < 4; i++) {
        for (int j = 0; j < 10; j++) {
            Value value(i, j);
            rings->add(i * 10l + j, &value, setInt);
        }
    }
    Print *print = new Print();
    rings->collect(0, 50, print, printData);

    fprintf(stdout, "Current Size: %d\n", rings->size());

    rings->expire(25);
    fprintf(stdout, "After Exipre Size: %d\n", rings->size());

    rings->collect(22, 35, print, printData);
}

void testBucketRingsAutoExpire() {
    BucketRingBuffers<Int> *rings = new BucketRingBuffers<Int>(20, 1000000);

    /**
     * Ring-0 0  [ 0,  9] 10 [10, 19]
     * Ring-1 20 [20, 29] 
     * Ring-2 30 [30, 39]
     */
    for (int i = 0; i < 4; i++) {
        for (int j = 1; j <= 10; j++) {
            Value value(i, j);
            rings->addAndExpire((i * 10l + j) * 1000000, 25, &value, setInt);
            fprintf(stdout, "Current Size: %d\n", rings->size());
        }
    }
    Print *print = new Print();
    rings->collect(23000000, 35000000, print, printData, getDataTime);
}

int main(int argc, char** argv) {
    testBucketRings();
    // testBucketRingsAutoExpire();
    return 0;
}