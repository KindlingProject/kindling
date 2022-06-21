include(ExternalProject)

set(PERF_SRC "${PROJECT_BINARY_DIR}/perf-prefix/src/perf")
message(STATUS "Using bundled perf in '${PERF_SRC}'")
set(PERF_LIB "${PERF_SRC}/tools/lib/perf/libperf.a")
set(PERF_API_LIB "${PERF_SRC}/tools/lib/api/libapi.a")
set(PERF_INCLUDE_DIR "${PERF_SRC}/tools/lib/perf/include")
ExternalProject_Add(perf
        PREFIX "${PROJECT_BINARY_DIR}/perf-prefix"
        URL "https://github.com/hocktea214/libperf/archive/refs/tags/v1.0.0.tar.gz"
        URL_HASH "SHA256=26c0c75194c87fe02b7bec7a64bd4c31ee1ee98e58a753e0cda4d6994591d05c"
        CONFIGURE_COMMAND ""
        BUILD_COMMAND ${CMD_MAKE}
        BUILD_BYPRODUCTS ${PERF_LIB}
        BUILD_IN_SOURCE 1
        INSTALL_COMMAND mkdir -p ${PERF_INCLUDE_DIR}/linux && cp ${PERF_SRC}/tools/include/uapi/linux/perf_event.h ${PERF_INCLUDE_DIR}/linux)
include_directories(${PERF_INCLUDE_DIR})