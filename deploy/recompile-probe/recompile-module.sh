mkdir rebuild-kindling-agent
cd rebuild-kindling-agent
curl -O https://k8s-bpf-probes-public.oss-cn-hangzhou.aliyuncs.com/2b4b2a107c05bd16a03c12841c6cce1d6404efac.tar.gz
tar -zvxf 2b4b2a107c05bd16a03c12841c6cce1d6404efac.tar.gz
cd agent-libs-2b4b2a107c05bd16a03c12841c6cce1d6404efac
docker pull kindlingproject/kernel-builder:latest
docker run -it -v /usr:/host/usr -v /lib/modules:/host/lib/modules -v $PWD:/source kindlingproject/kernel-builder:latest
cd ..
curl -O https://k8s-bpf-probes-public.oss-cn-hangzhou.aliyuncs.com/recompile-module-Dockerfile
mv -f recompile-module-Dockerfile Dockerfile
docker build -t kindlingproject/kindling-probe:latest-bymyself .

