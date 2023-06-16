FROM grafana/grafana:8.5.26
COPY topo-plugin /var/lib/grafana/plugins/kindlingproject-topology-panel
COPY docker/grafana.ini /etc/grafana
COPY docker/dashboards.yml /etc/grafana/provisioning/dashboards
COPY dashboard-json /etc/grafana/dashboards-files
