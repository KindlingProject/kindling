#docker build -f deploy/grafana-with-kindlingplugin/Dockerfile -t kindling-grafana:{version} .
FROM grafana/grafana:latest
COPY topo-plugin /var/lib/grafana/plugins/kindlingproject-topology-panel
COPY docker/grafana.ini /etc/grafana
COPY docker/dashboards.yml /etc/grafana/provisioning/dashboards
COPY dashboard-json /etc/grafana/dashboards-files