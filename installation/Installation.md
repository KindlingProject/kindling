# Quick Start
## Requirements
### Kubernetes
Kubernetes v1.x+ is required.
### Promethues
Deployed in Kubernetes.
### Grafana
Grafana is aready installed.
### Architecture
Only x86-64 Supported.
### Operating System
#### Distributions
Only supported on Linux. The following list describes the distribution versions (only official distribution, no kernel version changed) verified by Kindling. [More support list click here.](http://www.kindling.space:33215/project-1/doc-14/)

| **Distribution** | **Version (NO KENENL VERSION CHANGED)** |
| --- | --- |
| Ubuntu | 16.04+ |
| Debian | 9.0+ |
| RHEL | 7+ |
| CentOS | 7+ |

If Linux versions in your Kubernetes are **not listed**, you can compile with the steps in [Build Kindling container](#build-kindling-container) depending on your kernel version.
It's worth noting that the actual kernel version of some of Linux distributions may be inconsistent when the kernel is slightly updated such as an LTS kernel release. Some distributions such as Ubuntu, Debian may fail to load eBPF program, if so, try to set envrionment variable KERNEL_VERSION_CODE, which equals to (VERSION * 65536) + (PATCHLEVEL * 256) + SUBLEVEL.

#### Kernel Option

- The following kernel options must be enabled (usually they are, unless a custom-built kernel is used): 
   - `CONFIG_TRACEPOINTS`
   - `CONFIG_HAVE_SYSCALL_TRACEPOINTS`
- Furthermore, for **eBPF support**, the following **additional** kernel options must be enabled:
   - `CONFIG_BPF=y`, `CONFIG_BPF_JIT=y`, and `CONFIG_BPF_SYSCALL=y`
## Start
Kindling provides [script](https://raw.githubusercontent.com/Kindling-project/kindling/main/deploy/install.sh) and yaml files to deploy in Kubernetes:
```bash
# Make sure having the access to api-server, can run on Kubernetes master node.
bash <(curl -Ss https://raw.githubusercontent.com/Kindling-project/kindling/main/deploy/start.sh)
```
> **The url may not be accessed because kindling repository is private.**
> **Get the code first and 'cd kindling/deploy & bash install.sh'**

1. **Install Kindling Probe and Collector: install.sh** creates Namespace, ConfigMap, ClusterRole, ServiceAccount, ClusterRoleBindding. **install.sh** deploys the agent as two separate containers named **Kindling Probe** and **Kindling Collector**, which are combined in one Kubernetes Pod. The container images, namely **kindling-probe** and **kindling-collector**, are provided in Docker Hub, which can be **replaced with your own containers** in kindling-deploy.yml.
2. **Configure Promethues**: **install.sh** create **Service and Promethues ServiceMonitor** for exposing service to Promethues. You can refer to [How to use ServiceMonitor to scrape metric from Kindling](http://www.kindling.space:33215/project-1/doc-7/) for more information.
3. **Configure Grafana**: After the **install.sh** execves, you should config **grafana-plugins**, refer to [How to use grafana-plugin](http://www.kindling.space:33215/project-1/doc-8/).

Enjoy kindling!
# Build Kindling container
## Build kindling-probe

### Build local kernel modules and eBPF modules(Could skip if using precompiled modules)

Following steps are used to compile local kernel modules and eBPF modules **when your kernel version is not supported by Kindling**, which you can skip if using precompiled modules.

```bash
git clone -b kindling-dev https://github.com/Kindling-project/agent-libs
cd agent-libs
```

```bash
# Kernel headers are used to compile kernel modules and eBPF modules. The version of kernel headers must match the runtime. Warning: The command might not work with some kernel, or install kernel headers in another way. http://rpm.pbone.net is a choice to find RPMs for RHEL-like distributions.
# Debian-like distributions
sudo apt-get -y install linux-headers-$(uname -r)
# RHEL-like distributions
sudo yum -y install kernel-devel-$(uname -r)

# build and package eBPF, kernel probes
docker run -it -v /usr:/host/usr -v /lib/modules:/host/lib/modules -v $PWD:/source kindlingproject/kernel-builder:latest
tar -cvzf kindling-falcolib-probe.tar.gz kindling-falcolib-probe/
# copy and wait for building the image.You may need to prefix /kindling path with your own absolute path
cp kindling-falcolib-probe.tar.gz kindling/probe/deploy/
```

### Compile and build kindling-probe itself
```bash
git clone https://github.com/Kindling-project/kindling.git 
cd kindling/probe
```

```bash
# start compile container for binaries
./scripts/run_docker.sh
# or start in daemon mode, choose one of them
./scripts/run_docker_bpf_daemon.sh
# in container, compile kindling-probe
bazel build -s --config=clang src/probe:kindling_probe

# build container
# configure the image registry, repository and tag in probe/src/probe/BUILD.bazel
# If you use the kernel module and eBPF probe compiled in the previous step, execute the following command
bazel build -s --config=clang src/probe:push_image_localdriver
# else if you use kernel modules precompiled by kindling, execute the following command
bazel build -s --config=clang src/probe:push_image

# make sure you have access to push, push container image
./bazel-bin/src/probe/push_image_localdriver
# or
./bazel-bin/src/probe/push_image
```

## Build kindling-collector

```bash
git clone https://github.com/Kindling-project/kindling.git 

cd kindling/collector
docker run -it -v $PWD:/collector kindlingproject/kindling-collector-builder bash
go build -o kindling-collector
# exit from container
docker build -t kindling-collector -f deploy/Dockerfile .
# push container image by docker push
```

