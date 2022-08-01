#include <linux/kvm.h>
#include <fcntl.h>
#include <sys/ioctl.h>
#include <linux/perf_event.h>
#include <perf/evlist.h>
#include <perf/evsel.h>
#include <perf/cpumap.h>
#include <perf/threadmap.h>
#include <perf/mmap.h>
#include <perf/core.h>
#define _GNU_SOURCE
#include <stdio.h>
#include <unistd.h>
#include <sys/time.h>
#include "profile/perf/perf.h"

static int libperf_print(enum libperf_print_level level,
                         const char *fmt, va_list ap)
{
    return fprintf(stderr, fmt, ap);
}

static uint64_t time_ms(void)
{
    struct timeval tv;
    gettimeofday(&tv, NULL);
    return tv.tv_sec * 1000LL + tv.tv_usec / 1000;
}

int perf(struct perfData *data) {
    if (data->running) {
        //fprintf(stdout, "Perf has already started, skip it...\n");
        return 0;
    }

    struct perf_evlist *evlist;
	struct perf_evsel *evsel;
	struct perf_cpu_map *cpus;
	struct perf_event_attr attr = {
        .size           = sizeof(struct perf_event_attr),
		.type           = PERF_TYPE_SOFTWARE,
		.config         = PERF_COUNT_SW_CPU_CLOCK,
        .sample_period  = data->sampleMs*1000000,
		.sample_type    = PERF_SAMPLE_TID|PERF_SAMPLE_TIME|PERF_SAMPLE_CALLCHAIN,
        .disabled       = 1,
        .wakeup_events  = 1,
        .exclude_kernel = 0,
        .exclude_user   = 0,
        .use_clockid    = 1,
        .clockid        = 0, // CLOCK_REALTIME
	};
    uint64_t time_end;
    int time_left;
    int err = -1;

	libperf_init(libperf_print);
	cpus = perf_cpu_map__new(NULL);
	if (!cpus) {
		//fprintf(stderr, "failed to create cpus\n");
		return -1;
	}

	evlist = perf_evlist__new();
	if (!evlist) {
		//fprintf(stderr, "failed to create evlist\n");
		goto out_cpus;
	}

	evsel = perf_evsel__new(&attr);
	if (!evsel) {
		//fprintf(stderr, "failed to create cycles\n");
		goto out_cpus;
	}

    perf_evlist__add(evlist, evsel);
	perf_evlist__set_maps(evlist, cpus, NULL);
	err = perf_evlist__open(evlist);
	if (err) {
		//fprintf(stderr, "failed to open evlist\n");
		goto out_evlist;
	}

	err = perf_evlist__mmap(evlist, 4);
	if (err) {
		//fprintf(stderr, "failed to mmap evlist\n");
		goto out_evlist;
	}

	perf_evlist__enable(evlist);
    
    data->running = 1;

    //fprintf(stdout, "Start Perf...\n");
    time_end = time_ms() + data->collectMs;
    time_left = data->collectMs;
    while (data->running) {
        struct perf_mmap *map;
        union perf_event *event;
        int fds;

        fds = perf_evlist__poll(evlist, time_left);
        if (fds) {
            perf_evlist__for_each_mmap(evlist, map, false) {
                if (perf_mmap__read_init(map) < 0)
			        continue;
                while ((event = perf_mmap__read_event(map)) != NULL) {
                    if (event->header.type == PERF_RECORD_SAMPLE && data->sample) {
                        struct sample_type_data *sample_data = (void *)event->sample.array;
                        data->sample(sample_data);
                    }
                    perf_mmap__consume(map);
                }
                perf_mmap__read_done(map);
            }
        }
        if (time_left == 0) {
            data->collect();
        }            

        time_left = time_end - time_ms();
        if (time_left <= 0) {
            time_end = time_ms() + data->collectMs + time_left;
            time_left = 0;
        }
    }

    perf_evlist__disable(evlist);
    perf_evlist__munmap(evlist);
out_evlist:
	perf_evlist__delete(evlist);
out_cpus:
	perf_cpu_map__put(cpus);

    //fprintf(stdout, "End Perf...\n");
	return err;
}
