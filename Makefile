GitCommit=${shell git rev-parse --short HEAD || echo unknow}

define exec-command
$(1)
endef

.PHONY: collector
collector: libkindling
	@echo Build Env:
	@echo Agent is build based on commit: ${GitCommit}
	cd collector && go build -o docker/kindling-collector -ldflags="-X 'github.com/Kindling-project/kindling/collector/version.CodeVersion=${GitCommit}'" ./cmd/kindling-collector/

.PHONY: libkindling
libkindling:
	mkdir -p probe/build && cd probe/build && cmake -DBUILD_DRIVER=OFF -DPROBE_VERSION=0.1.1dev .. && make 
	mkdir -p collector/docker/libso/. && cp -rf probe/build/src/libkindling.so collector/docker/libso/. && cp -rf probe/build/src/libkindling.so /usr/lib64/.



LIBS_SRC := $(abspath ./probe/libs/agent-libs)
COLLECTOR_PATH := $(abspath ./collector)
SCRIPTS_PATH := $(abspath ./deploy/scripts)
PROBE_PATH := $(abspath ./probe)
PROBE_PATCH_FILE := $(PROBE_PATH)/cmake/modules/agent-libs.cmake

kernelVersion = $(shell uname -r)
ErrorExit = exit 1

ifneq ($(shell cat /etc/os-release | grep Ubuntu | wc -l),0)
	KernelName = linux-headers-$(kernelVersion)
	CheckCommand = dpkg -l | grep $(KernelName)
	InstallCommand = sudo apt-get -y install $(KernelName)
else
	ifneq ($(shell cat /etc/os-release | grep CentOS | wc -l),0)
		KernelName = kernel-devel-$(kernelVersion)
		CheckCommand = yum list installed |grep $(KernelName)
		InstallCommand = sudo yum -y install $(KernelName)
	else
		KernelName = NotSupport
		CheckCommand = $(ErrorExit)
		InstallCommand = $(ErrorExit)
	endif
endif

LIBS_IMAGE ?= kindlingproject/kernel-builder:latest
KINDLING_IMAGE ?= kindlingproject/agent-builder:latest


## Check if probe build image and collector build image exists or pull from dockerhub
.PHONY: builder-images
builder-images:
	@echo "Checking builder images...";
	@if [ -z "$(shell docker images -q $(LIBS_IMAGE))" ]; then \
		echo "Downloading probe build image..."; \
		docker pull $(LIBS_IMAGE); \
	else \
		echo "Probe build image already exists"; \
	fi
	@if [ -z "$(shell docker images -q $(KINDLING_IMAGE))" ]; then \
		echo "Downloading collector build image..."; \
		docker pull $(KINDLING_IMAGE); \
	else \
		echo "Collector build image already exists"; \
	fi


## Check kernel headers exists or download, support Ubuntu, CentOS
.PHONY: kernel-headers
kernel-headers:
	@echo "Checking kernel headers...";
	@if [ "$(KernelName)" = "NotSupport" ]; then \
		echo "System kernel is not the expected Ubuntu and CentOS.Try installing kernel headers locally"; \
		$(ErrorExit); \
	fi

	@if [ -z "$(shell exit 1)" ]; then \
		echo "Downloading $(KernelName)..."; \
		$(InstallCommand); \
	else \
		echo "$(KernelName) already installed"; \
	fi

## Check libs exists or download, patch cmake file
.PHONY: patch-libs
patch-libs:
	@echo "Checking libs...";
	@if [ ! -d "$(LIBS_SRC)" ]; then \
		echo "$(LIBS_SRC) not exist, downloading....."; \
		git submodule update --init --recursive; \
	else \
		echo "Libs src: $(LIBS_SRC) exist"; \
	fi

## Grant executable permissions to build.sh
.PHONY: patch-scripts
patch-scripts:
	@echo "Grant executable permissions to build.sh";
	@cd $(SCRIPTS_PATH) && chmod +x build.sh;

## Check all depends
.PHONY: depends
depends: builder-images kernel-headers patch-libs patch-scripts

## Build probe in docker with /lib/modules
.PHONY: build-probe
build-probe:
	@echo "Building kindling probe..."
	@docker run \
		--env "ENV_HEADER_VERSION=$(kernelVersion)" \
		--rm -it \
		-v /usr:/host/usr \
		-v /lib/modules:/host/lib/modules \
		-v $(LIBS_SRC):/source \
		$(LIBS_IMAGE);

.PHONY: pack-probe
pack-probe:
	@echo "Packaging kindling probe...";
	@if [ ! -d "$(LIBS_SRC)/kindling-falcolib-probe/" ]; then \
		echo "The packaged probe does not exist.try again"; \
		exit 1; \
	else \
		cd $(LIBS_SRC) && \
		tar -cvzf kindling-falcolib-probe.tar.gz kindling-falcolib-probe/ && \
		mv kindling-falcolib-probe.tar.gz $(COLLECTOR_PATH)/docker/ ;\
	fi

## Build kindling probe in docker
.PHONY: kindling-probe
kindling-probe: build-probe pack-probe

.PHONY: _build-collector
_build-collector:
	@cd $(SCRIPTS_PATH) && sh ./build.sh

## Build kindling collector in docker
.PHONY: kindling-collector
kindling-collector:
	@if [ ! -f "$(COLLECTOR_PATH)/docker/kindling-falcolib-probe.tar.gz" ]; then \
		echo "The kindling-falcolib-probe.tar.gz does not exist. make probe-libs first"; \
		exit 1; \
	fi
	@cd $(SCRIPTS_PATH) && bash -c "./run_docker.sh make _build-collector";
	@exit ;

## Install depends and build probe locally and build collector
.PHONY: docker-build-all
docker-build-all: depends kindling-probe kindling-collector