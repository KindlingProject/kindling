## 编译
---
### 启动编译镜像
```shell
cd scripts
./run_docker.sh

### daemon形式启动
./run_docker_bpf_daemon.sh
```
### 编译probe二进制
```shell
bazel build -s --config=clang src/probe:kindling_probe
```

### 编译运行镜像
```shell
### 编译
bazel build -s --config=clang src/probe:push_image
### 推送，这里需要registry的密码，可以修改 src/probe/BUILD.bazel 来修改推送的registry
./bazel-bin/src/probe/push_image
```