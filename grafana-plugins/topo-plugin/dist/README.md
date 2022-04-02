## This Topology Plugin is desgin for project Kindling?
This is a topology component based on ANTV developed specifically for the Kindling project. Prometheus data is queried using modified EBPF probes, so this topology is not a generic plugin。
To use this component together, you need to import a kindling customized dashboard([dashboard.json](https://github.com/Kindling-project/kindling/blob/main/grafana-plugins/dashboard-json/topology.json)) that goes back to Prometheus to query the data collected by the probe
If you want to use [Kindling](https://github.com/Kindling-project/kindling)


After the integration plugin, configure the corresponding Prometheus data source and you will see the plugin as shown
![img](https://raw.githubusercontent.com/thousandxu/zipImage/main/topo-plugin/topo.png)

## What does this plugin do？
Topology call relationships are generated based on data queried in Prometheus and can be aggregated based on namespace and workload. The topology call relationship displays indicators such as Latency、calls、error rate and Volume. You can also view indicator data of nodes。
