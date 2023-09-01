# README
该仓库只包含`Trace Profiling`功能的agent代码，不包含前端代码和后端代码。

## 命名
仓库命名为`camera-agent`的原因：
- 该仓库在`Kindling`组织下且为私有仓库，不需要增加kindling供外部了解项目所属组织；
- 该仓库只包含`Trace Profiling`功能的agent代码，不包含前端代码和后端代码，所以使用`agent`尾缀；
- 没有使用`rust`字眼是因为未来只会存在一个`Trace Profiling`探针，该探针使用Rust和C/C++实现，存在`camera`标识的情况下不需要使用`rust`标识不会造成混淆；

## 如何进行Profiling
使用[flamegraph-rs](https://github.com/flamegraph-rs/flamegraph)生成火焰图

### 1 安装FlameGraph
> cargo install flamegraph

### 2 开启Cargo.toml配置
```
[profile.release]
debug = true
```
### 3 编译二进制
> cargo build --release

### 4 启动二进制
```
cp target/release/camera-agent /app
cd /app
./camera-agent
```

### 5 生成火焰图
```
# 找到PID
ps -ef|grep camera-agent

# 生成火焰图 flamegraph.svg
flamegraph --pid 1337
```