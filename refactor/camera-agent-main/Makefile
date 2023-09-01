LIBS_SRC := $(abspath ./src/probe/libs/agent-libs)

.PHONY: collector
collector: libkindling
	cargo build --profile=release
	cp ./target/release/camera-agent docker/

.PHONY: libkindling
libkindling: patch-libs
	mkdir -p src/probe/build && cd src/probe/build && cmake -DBUILD_DRIVER=OFF -DAGENT_LIBS_SOURCE_DIR=$(LIBS_SRC) .. && make 
	mkdir -p docker/libso/. && cp -rf src/probe/build/src/libkindling.so docker/libso/. && cp -rf src/probe/build/src/libkindling.so /usr/lib64/.

## Check libs exists or download, patch cmake file
.PHONY: patch-libs
patch-libs:
	@echo "Checking libs...";
	@if [ ! "$(shell ls -A1q $(LIBS_SRC))" ]; then \
		echo "$(LIBS_SRC) not exist, downloading....."; \
		git submodule update --init --recursive; \
	else \
		echo "Libs src: $(LIBS_SRC) exist"; \
	fi