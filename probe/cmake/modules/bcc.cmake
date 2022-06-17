include(ExternalProject)
project(. NONE)

set(BCC_SRC "${PROJECT_BINARY_DIR}/bcc-prefix/src/bcc")
message(STATUS "Using bundled bcc in '${BCC_SRC}'")
set(BCC_LIB "${BCC_SRC}/src/cc/libbcc.a")
set(BCC_INCLUDE_DIR "${BCC_SRC}/src/cc")
ExternalProject_Add(bcc
        PREFIX "${PROJECT_BINARY_DIR}/bcc-prefix"
        URL "https://github.com/hocktea214/bcc/archive/refs/tags/v0.24.0-libbcc.tar.gz"
        URL_HASH "SHA256=d8a87a4cee5332eacdfdf40428b0f1f8b0236e600dbce56d917b311076c5e905"
        BUILD_COMMAND ${CMD_MAKE}
        BUILD_BYPRODUCTS ${BCC_LIB}
        BUILD_IN_SOURCE 1
        INSTALL_COMMAND "")

include_directories(${BCC_INCLUDE_DIR})