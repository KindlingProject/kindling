mkdir rebuild-kindling-agent
cd rebuild-kindling-agent
curl -O https://k8s-bpf-probes-public.oss-cn-hangzhou.aliyuncs.com/66fe2a5d4cfd2566108e2519b46a70fb4a247741.tar.gz
tar -zvxf 66fe2a5d4cfd2566108e2519b46a70fb4a247741.tar.gz
cd agent-libs-66fe2a5d4cfd2566108e2519b46a70fb4a247741
docker pull kindlingproject/kernel-builder:latest
docker run -it -v /usr:/host/usr -v /lib/modules:/host/lib/modules -v $PWD:/source kindlingproject/kernel-builder:latest
cd ..
curl -O https://k8s-bpf-probes-public.oss-cn-hangzhou.aliyuncs.com/recompile-module-Dockerfile-0.3.0
mv -f recompile-module-Dockerfile-0.3.0 Dockerfile
docker build -t kindlingproject/kindling-probe:latest-bymyself .

