FROM centos:7

WORKDIR /app/
RUN  curl https://k8s-bpf-probes-public.oss-cn-hangzhou.aliyuncs.com/kindling-falcolib-probe-v0.4.0.tar.gz -o kindling-falcolib-probe.tar.gz
COPY libso/libkindling.so /lib64/
RUN ldconfig

COPY kindling-probe-loader /usr/bin/kindling-probe-loader
RUN chmod +x  /usr/bin/kindling-probe-loader
COPY kindling-collector-config.yml /etc/kindling/config/
COPY kindling-collector /usr/bin/kindling-collector
COPY start.sh /app/

CMD ["sh", "start.sh"]
