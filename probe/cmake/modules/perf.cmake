include(ExternalProject)
project(. NONE)

set(PERF_SRC "${PROJECT_BINARY_DIR}/perf-prefix/src/perf")
message(STATUS "Using bundled perf in '${PERF_SRC}'")
set(PERF_LIB "${PERF_SRC}/tools/lib/perf/libperf.so")
set(PERF_INCLUDE_DIR "${PERF_SRC}/tools/lib/perf/include")
ExternalProject_Add(perf
        PREFIX "${PROJECT_BINARY_DIR}/perf-prefix"
        URL "https://github.com/hocktea214/libperf/archive/refs/tags/v1.0.0.tar.gz"
        URL_HASH "SHA256=ed38b724e8986b114f0b3214d552c9e2608ff262eedfe6858a98aebec952abae"
        CONFIGURE_COMMAND ""
        BUILD_COMMAND ${CMD_MAKE}
        BUILD_BYPRODUCTS ${PERF_LIB}
        BUILD_IN_SOURCE 1
        INSTALL_COMMAND "")

include_directories(${PERF_INCLUDE_DIR})