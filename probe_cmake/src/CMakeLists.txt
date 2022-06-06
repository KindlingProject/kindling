
include(libscap)
include(libsinsp)

include_directories(.)
include_directories("${LIBSINSP_INCLUDE_DIRS}")
include_directories("../util")

set(SOURCE_FILES
		cgo/cgo_func.cpp
		util/util.cpp)

add_library(kindling SHARED ${SOURCE_FILES})
set(CMAKE_POSITION_INDEPENDENT_CODE ON)

target_link_libraries(kindling
		sinsp)
install(TARGETS kindling
		DESTINATION bin)