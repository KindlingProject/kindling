GitCommit=${shell git rev-parse --short HEAD || echo unknow}

define exec-command
$(1)
endef

.PHONY: collector
collector: libkindling
	@echo Build Env:
	@echo Agent is build based on commit: ${GitCommit}
	cd collector && go build -o docker/kindling-collector -ldflags="-X 'github.com/Kindling-project/kindling/collector/version.CodeVersion=$GitCommit'" ./cmd/kindling-collector/

.PHONY: libkindling
libkindling:
	mkdir -p probe/build && cd probe/build && cmake -DBUILD_DRIVER=OFF -DPROBE_VERSION=0.1.1dev .. && make 
	mkdir -p collector/docker/libso/. && cp -rf probe/build/src/libkindling.so collector/docker/libso/. && cp -rf probe/build/src/libkindling.so /usr/lib64/.