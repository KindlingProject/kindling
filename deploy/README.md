# Deployment files for Kubernetes
The files under this directory are needed when releasing the Kindling.

## agent
This directory contains files used for deploying `kindling-agent` in Kubernetes. 

## grafana-with-plugins
This directory contains files used for deploying `Grafana` in Kubernetes. 

## recompile-probe
 The files under this directory provide a convenient way to build a new `kindling-probe` image with `drivers` **built locally**. The script compiles the driver codes and produces `drivers` based on the specific kernel version, and rebuilds a container image base on the `drivers` and the latest `kindling-probe` image. 