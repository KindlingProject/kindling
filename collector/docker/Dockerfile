FROM centos:7

COPY libso/libzmq.so.5 /lib64/
RUN ldconfig

COPY kindling-collector-config.yml /etc/kindling/config/
COPY kindling-collector /usr/bin/kindling-collector

CMD ["/usr/bin/kindling-collector", "--config=/etc/kindling/config/kindling-collector-config.yml"]
