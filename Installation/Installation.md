# Quick Start
## Requirements
### Kubernetes
Kubernetes v1.x+ is required.
### Promethues
Deployed in Kubernetes.
### Grafana
TBD
### Architecture
Only x86-64 Supported.
### Operating System
#### Distributions
Only supported on Linux. The following list describes the distribution versions (only official distribution, no kernel version changed) verified by Kindling. [More support list click here.](./Distributions and Kernel Support List.md)

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
bash <(curl -Ss https://raw.githubusercontent.com/Kindling-project/kindling/main/deploy/install.sh)
```
> **The url may not be accessed because kindling repository is private.**
> **Get the code first and 'cd kindling/deploy & bash install.sh'**

1. **Install Kindling Probe and Collector: install.sh** creates Namespace, ConfigMap, ClusterRole, ServiceAccount, ClusterRoleBindding. **install.sh** deploys the agent as two separate containers named **Kindling Probe** and **Kindling Collector**, which are combined in one Kubernetes Pod. The container images, namely **kindling-probe** and **kindling-collector**, are provided in Docker Hub, which can be** replaced with your own containers** in kindling-deploy.yml.
1. **Configure Promethues**: **install.sh **create** Service **and** Promethues ServiceMonitor **for exposing service to Promethues**. You can refer to **[How to use ServiceMonitor to scrape metric from Kindling](./How to use ServiceMonitor to scrape metric from Kindling.md ) for more information.
3. **Configure Grafana**: After the **install.sh** execves, you should config **grafana-plugins**, refer to [How to use grafana-plugin](./How to  use grafana-plugin.md).

Enjoy kindling!
# Build Kindling container
## Compile in Docker
### Install target kernel headers
Kernel headers are used to compile kernel module and eBPF module. The version of kernel headers must match the runtime.
```
# Debian-like distributions
sudo apt-get -y install linux-headers-$(uname -r)
# RHEL-like distributions
sudo yum -y install kernel-devel-$(uname -r)
```
**Warning**: The command might not work with some kernel, or install kernel headers in another way. [http://rpm.pbone.net](http://rpm.pbone.net) is a choice to find RPMs for RHEL-like distributions.
### Build Container
Kindling provides a container environment for compiling, run as follows:
```bash
git clone https://github.com/Kindling-project/kindling.git 
cd kindling

# start compile container.
./scripts/run_docker.sh
# or start in daemon mode
./scripts/run_docker_bpf_daemon.sh

# build kindling-probe
cd probe
# compile kindling-probe
bazel build -s --config=clang src/probe:kindling_probe
# build container
bazel build -s --config=clang src/probe:push_image
./bazel-bin/src/probe/push_image

# build kindling-collector
cd collector
go build
cd deploy
docker build -t kindling-collector -f deploy/Dockerfile .
```
