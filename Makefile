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



LIBS_SRC := $(abspath ./agent-libs)
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

.PHONY: builder-images
builder-images:
	@echo "Check dependencies images...";
	@if [ -z "$(shell docker images -q $(LIBS_IMAGE))" ]; then \
		echo "Downloading probe build image..."; \
		docker pull $(LIBS_IMAGE); \
	else \
		echo "Probe build image already exists"; \
	fi
	@if [ -z "$(shell docker images -q $(KINDLING_IMAGE))" ]; then \
		echo "Downloading kindling build image..."; \
		docker pull $(KINDLING_IMAGE); \
	else \
		echo "Kindling build image already exists"; \
	fi

.PHONY: kernel-headers
kernel-headers:
	@echo "Checking for kernel-headers...";
	@if [ "$(KernelName)" = "NotSupport" ]; then \
  		echo "$(KernelName) install kernel header,try install local"; \
		$(ErrorExit); \
	fi

	if [ -z "$(shell exit 1)" ]; then \
		echo "Downloading $(KernelName)..."; \
		$(InstallCommand); \
    	else \
    	echo "$(KernelName) already installed"; \
    fi

AGENT_LIBS_COMMIT_ID=$(shell cd $(LIBS_SRC) && git rev-parse HEAD)

.PHONY: patch-libs
patch-libs:
	@if [ ! -d "$(LIBS_SRC)" ]; then \
  		echo "$(LIBS_SRC) not exist, download....."; \
		git submodule update --init --recursive; \
	else \
	  echo "agent-libs: $(LIBS_SRC) exist"; \
	fi
	@echo "Fetching agent-libs commit id: $(AGENT_LIBS_COMMIT_ID)";

	@echo "patching $(PROBE_PATCH_FILE)";
	@sed 's/AGENT_LIBS_VERSION "[0-9a-zA-Z]*"/AGENT_LIBS_VERSION "${AGENT_LIBS_COMMIT_ID}"/g' $(PROBE_PATCH_FILE) > $(PROBE_PATCH_FILE).new && \
	mv $(PROBE_PATCH_FILE).new $(PROBE_PATCH_FILE);
	@echo "patching $(PROBE_PATCH_FILE) done"

.PHONY: patch-scripts
patch-scripts:
	@cd $(SCRIPTS_PATH) && chmod +x build.sh;


.PHONY: dependencies
dependencies: builder-images kernel-headers patch-libs patch-scripts

.PHONY: build-libs
build-libs: ## build libs in docker with /lib/modules
	@echo "Building kindling libs..."
	@docker run \
		--env "ENV_HEADER_VERSION=$(kernelVersion)" \
		--rm -it \
		-v /usr:/host/usr \
		-v /lib/modules:/host/lib/modules \
		-v $(LIBS_SRC):/source \
		$(PROBE_IMAGE);

.PHONY: pack-libs
pack-libs:
	@echo "Packaging kindling libs...";
	@if [ ! -d "$(LIBS_SRC)/kindling-falcolib-probe/" ]; then \
    		echo "The packaged probe does not exist.try again"; \
    		exit 1; \
    else \
		cd $(LIBS_SRC) && \
		tar -cvzf kindling-falcolib-probe.tar.gz kindling-falcolib-probe/ && \
		cp -f kindling-falcolib-probe.tar.gz $(COLLECTOR_PATH)/docker/.; \
	fi

.PHONY: agent-libs
agent-libs: build-libs pack-libs

.PHONY: build-collector
build-collector:
	@cd $(SCRIPTS_PATH) && sh ./build.sh

.PHONY: kindling-collector
kindling-collector: ## build kindling collector in docker
	@if [ ! -f "$(COLLECTOR_PATH)/docker/kindling-falcolib-probe.tar.gz" ]; then \
			echo "The kindling-falcolib-probe.tar.gz does not exist. make probe-libs first"; \
			exit 1; \
	fi
	@cd $(SCRIPTS_PATH) && bash -c "./run_docker.sh make build-collector";
	@exit ;

.PHONY: docker-build-all
docker-build-all: dependencies agent-libs kindling-collector