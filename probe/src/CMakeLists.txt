
include(libscap)
include(libsinsp)

include_directories(.)
include_directories("${LIBSINSP_INCLUDE_DIRS}")
include_directories("../cgo")
include_directories("../converter")
set(SOURCE_FILES
		cgo/cgo_func.cpp
		cgo/kindling.cpp
		converter/cpu_converter.cpp
        cgo/catch_sig.cpp cgo/catch_sig.h)

add_library(kindling SHARED ${SOURCE_FILES})
set(CMAKE_POSITION_INDEPENDENT_CODE ON)

target_link_libraries(kindling
		sinsp
		Qt5::Core)
install(TARGETS kindling
		DESTINATION bin)