AGENT_LIBS_COMMIT_ID=1daf5aa0e534cd96f9801729df35e5942c05f2bd

mkdir -p rebuild-kindling-agent
cd rebuild-kindling-agent
curl -O https://k8s-bpf-probes-public.oss-cn-hangzhou.aliyuncs.com/agent-libs-${AGENT_LIBS_COMMIT_ID}.tar.gz
tar -zvxf agent-libs-${AGENT_LIBS_COMMIT_ID}.tar.gz
cd agent-libs-${AGENT_LIBS_COMMIT_ID}
docker pull kindlingproject/kernel-builder:latest
docker run -it -v /usr:/host/usr -v /lib/modules:/host/lib/modules -v $PWD:/source kindlingproject/kernel-builder:latest
cd ..
cat <<EOF > Dockerfile
FROM kindlingproject/kindling-agent:latest
COPY ./agent-libs-${AGENT_LIBS_COMMIT_ID}/kindling-falcolib-probe/* /opt/kindling-extra-probe/
EOF
docker build -t kindlingproject/kindling-agent:latest-bymyself .