FROM centos:7
RUN yum install -y gdb
ENV HCMINE_HOST_ROOT /host

COPY lib/* /lib64/
RUN ldconfig

COPY kindling-probe-loader /usr/bin/kindling-probe-loader
RUN chmod +x /usr/bin/kindling-probe-loader
COPY docker-entrypoint.sh /
RUN chmod +x /docker-entrypoint.sh
COPY gdb_print.sh /usr/bin/gdb_print.sh
COPY probe/* /opt/.kindling/

COPY kindling-probe /usr/bin/

ENTRYPOINT ["/docker-entrypoint.sh"]

CMD ["/usr/bin/kindling-probe"]