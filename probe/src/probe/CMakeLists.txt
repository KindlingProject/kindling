
add_compile_options(-fPIC)
add_compile_options(-pie)
cmake_minimum_required(VERSION 2.8.2)

set(DEBUG_FLAGS "-O0 -D_DEBUG")
set(CMAKE_COMMON_FLAGS "-Wall -ggdb")

if (BUILD_WARNINGS_AS_ERRORS)
    set(CMAKE_SUPPRESSED_WARNINGS "-Wno-unused-parameter -Wno-missing-field-initializers -Wno-sign-compare -Wno-type-limits -Wno-implicit-fallthrough -Wno-format-truncation")
    set(CMAKE_COMMON_FLAGS "${CMAKE_COMMON_FLAGS} -Wextra -Werror ${CMAKE_SUPPRESSED_WARNINGS}")
endif ()

set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} ${CMAKE_COMMON_FLAGS}")
set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} ${CMAKE_COMMON_FLAGS} -std=c++0x")

set(CMAKE_C_FLAGS_DEBUG "${DEBUG_FLAGS}")
set(CMAKE_CXX_FLAGS_DEBUG "${DEBUG_FLAGS}")

set(CMAKE_C_FLAGS_RELEASE "-O3 -fno-strict-aliasing -DNDEBUG")
set(CMAKE_CXX_FLAGS_RELEASE "-O3 -fno-strict-aliasing -DNDEBUG")

add_definitions(-DPLATFORM_NAME="${CMAKE_SYSTEM_NAME}")
add_definitions(-DK8S_DISABLE_THREAD)
add_definitions(-DHAS_CAPTURE)

include_directories(.)
include_directories("../driver")
include_directories("../userspace/libscap")
include_directories("../userspace/libsinsp")
include_directories("../converter")
include_directories("../publisher")
include_directories("${PROTOBUF_INCLUDE}")
include_directories("${JSONCPP_INCLUDE}")
include_directories("${TBB_INCLUDE_DIR}")
include_directories(${PROJECT_BINARY_DIR}/driver/src)

if(NOT WIN32)
	include_directories("${CURL_INCLUDE_DIR}")
	include_directories("${CURSES_INCLUDE_DIR}")
endif()
set(SOURCE_FILES
        converter/sysdig_converter.cpp
        converter/kindling_event.pb.cc
        publisher/subscribe.pb.cc
        publisher/publisher.cpp
        catch_sig.cpp
        main.cpp
        )

add_executable(kindling-probe ${SOURCE_FILES})
find_package(cppzmq) # TODO
set(CONNTRACK_LIB "/usr/local/lib/libnetfilter_conntrack.so")
target_link_libraries(kindling-probe
        sinsp
        cppzmq)

install(TARGETS kindling-probe
        DESTINATION /usr/bin)

install(FILES lib/libjq.so.1.0.4 lib/libjq.so.1
            lib/libprotobuf.so.15.0.0 lib/libprotobuf.so.15
            lib/libz.so.1.2.11 lib/libz.so.1
        DESTINATION /lib64)
