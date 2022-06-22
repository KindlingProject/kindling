
include(libscap)
include(libsinsp)

include_directories(.)
include_directories("${LIBSINSP_INCLUDE_DIRS}")
include_directories("../cgo")
include_directories("../converter")
include_directories("../log")
include_directories("../profile")
include_directories("../util")
set(SOURCE_FILES
		cgo/cgo_func.cpp
		cgo/kindling.cpp
		converter/cpu_converter.cpp
		converter/event_cache.cpp
		converter/epoll_event_cache.cpp
		log/log_info.cc
		profile/bcc/symbol.cc
		profile/perf/perf.c
		profile/flame_graph.cc
		profile/profiler.cc)

add_library(kindling SHARED ${SOURCE_FILES})
set(CMAKE_POSITION_INDEPENDENT_CODE ON)

target_link_libraries(kindling
		sinsp
		${PERF_LIB}
		${PERF_API_LIB}
		${BCC_LIB})
install(TARGETS kindling
		DESTINATION bin)