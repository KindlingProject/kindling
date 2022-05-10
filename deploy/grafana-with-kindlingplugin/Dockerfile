#docker build -f deploy/grafana-with-kindlingplugin/Dockerfile -t kindling-grafana:{version} .
FROM grafana/grafana:latest
COPY grafana-plugins/topo-plugin /var/lib/grafana/plugins/kindlingproject-topology-panel
COPY deploy/grafana-with-kindlingplugin/grafana.ini /etc/grafana
COPY deploy/grafana-with-kindlingplugin/dashboards.yml /etc/grafana/provisioning/dashboards
COPY grafana-plugins/dashboard-json /etc/grafana/dashboards-files