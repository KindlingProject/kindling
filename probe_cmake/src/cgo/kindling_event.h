#ifndef KINDLING_EVENT_H
#define KINDLING_EVENT_H

struct kindling_event_t_for_go {
    uint64_t timestamp;
    char *name;
    uint32_t category;
    struct key_value {
        char *key;
        char *value;
        uint32_t valueType;
    } userAttributes[8];
    struct thread_info {
        uint32_t pid;
        uint32_t tid;
        uint32_t uid;
        uint32_t gid;
        char *comm;
        char *containerId;
    } tinfo;
    struct fd_info {
        int32_t num;
        uint32_t fdType;

        char *filename;
        char *directory;

        uint32_t protocol;
        uint8_t role;
        uint32_t sip;
        uint32_t dip;
        uint32_t sport;
        uint32_t dport;

        uint64_t source;
        uint64_t destination;
    } fdInfo;
};

#endif //KINDLING_EVENT_H