FROM centos:7

WORKDIR /app/
RUN yum install -y qt5-qtbase-devel
RUN yum install -y gdb
COPY gdb_print.sh /app/
RUN curl https://k8s-bpf-probes-public.oss-cn-hangzhou.aliyuncs.com/kindling-falcolib-probe-v0.5.0.tar.gz -o kindling-falcolib-probe.tar.gz
RUN curl -O https://k8s-bpf-probes-public.oss-cn-hangzhou.aliyuncs.com/async-profiler-kindling-v1.0.1.tar.gz
RUN tar -zvxf async-profiler-kindling-v1.0.1.tar.gz

COPY libso/libkindling.so /lib64/
COPY libso/* /usr/lib64/

COPY kindling-probe-loader /usr/bin/kindling-probe-loader
RUN chmod +x  /usr/bin/kindling-probe-loader
COPY kindling-collector-config.yml /app/config/
COPY kindling-collector /app/kindling-collector
COPY libkindling-plugin.so /app/
COPY start.sh /app/

CMD ["sh", "start.sh"]
