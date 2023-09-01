FROM centos:7

WORKDIR /app/
RUN yum install -y qt5-qtbase-devel
RUN yum install -y ntpdate
RUN yum install -y gdb
COPY gdb_print.sh /app/
RUN curl https://k8s-bpf-probes-public.oss-cn-hangzhou.aliyuncs.com/kindling-falcolib-probe-latest.tar.gz -o kindling-falcolib-probe.tar.gz

COPY build-asyncprofiler.sh /app/
RUN sh build-asyncprofiler.sh

COPY start.sh /app/
COPY kindling-probe-loader /usr/bin/kindling-probe-loader
RUN chmod +x  /usr/bin/kindling-probe-loader

COPY libso/* /usr/lib64/
COPY libkindling-plugin.so /app/
COPY libso/libkindling.so /lib64/
COPY config/* /app/config/
COPY camera-agent /app/camera-agent

CMD ["sh", "start.sh"]