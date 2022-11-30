cmake_minimum_required(VERSION 3.5.1)

project(. NONE)

include(ExternalProject)
message(STATUS "Driver version: ${AGENT_LIBS_VERSION}")

ExternalProject_Add(
  agent-libs
  URL "https://github.com/CloudDectective-Harmonycloud/agent-libs/archive/${AGENT_LIBS_VERSION}.tar.gz"
  URL_HASH "${AGENT_LIBS_CHECKSUM}"
  CONFIGURE_COMMAND ""
  BUILD_COMMAND ""
  INSTALL_COMMAND ""
  TEST_COMMAND "")