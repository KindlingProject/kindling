cmake_minimum_required(VERSION 3.5.1)

project(. NONE)

include(ExternalProject)
message(STATUS "Driver version: ${AGENT_LIBS_VERSION}")

if (AGENT_LIBS_SOURCE_DIR)
    ExternalProject_Add(
            agent-libs
            SOURCE_DIR "${AGENT_LIBS_SOURCE_DIR}"
            CONFIGURE_COMMAND ""
            BUILD_COMMAND ""
            INSTALL_COMMAND ""
            TEST_COMMAND "")
else ()
    ExternalProject_Add(
            agent-libs
            URL "https://github.com/KindlingProject/agent-libs/archive/${AGENT_LIBS_VERSION}.tar.gz"
            URL_HASH "${AGENT_LIBS_CHECKSUM}"
            CONFIGURE_COMMAND ""
            BUILD_COMMAND ""
            INSTALL_COMMAND ""
            TEST_COMMAND "")
endif ()