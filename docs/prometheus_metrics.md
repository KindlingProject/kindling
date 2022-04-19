# Prometheus Metrics Description
## Service Metrics
Service metrics are generated from the server-side events, which are used to show the quality of service. 
| **Metric Name** | **Type** | **Description** |
| --- | --- | --- |
| kindling_entity_request_total | Counter | Total number of requests |
| kindling_entity_request_duration_nanoseconds_total | Counter | Total duration of requests |
| kindling_entity_request_send_bytes_total | Counter | Total size of payload sent |
| kindling_entity_request_receive_bytes_total | Counter | Total size of payload received |
| kindling_entity_request_average_duration_nanoseconds_count  | Counter(Histogram) | Count of average duration of requests |
| kindling_entity_request_average_duration_nanoseconds_sum | Counter(Histogram) | Sum of average duration of requests |
| kindling_entity_request_average_duration_nanoseconds_bucket | Counter(Histogram) | Histogram buckets of average duration of requests |

| **Label Name** | **Example** | **Notes** |
| --- | --- | --- |
| node | worker-1 | Node name represented in Kubernetes cluster |
| namespace | default | Namespace of the Pod |
| workload_kind | daemonset | K8sResourceType |
| workload_name | api-ds | K8sResourceName |
| service | api | One of the services that target this pod |
| pod | api-ds-xxxx | The name of the pod |
| container | api-container | The name of the container |
| container_id | 1a2b3c4d5e6f | The shorten container id which contains 12 characters |
| ip | 10.1.11.23 | The IP address of the entity |
| port | 80 | The listening port of the entity |
| protocol | http | The application layer protocol the requests use |
| request_content | /test/api | The request content of the requests |
| response_content | 200 | The response content of the requests |
| is_slow | false | (Only applicable to `kindling_entity_request_total`)Whether the requests are considered as slow |

**The labels "request_content" and "response_content" hold different values when "protocol" is different.**

- When protocol is http:
  
| **Label** | **Example** | **Notes** |
| --- | --- | --- |
| request_content | /test/api | Endpoint of HTTP request. URL has been truncated to avoid high-cardinality. |
| response_content | 200 | 'Status Code' of HTTP response. |

- When protocol is dns:
  
| **Label** | **Example** | **Notes** |
| --- | --- | --- |
| request_content | www.google.com | Domain to be queried |
| response_content | 0 | "rcode" of DNS response. Including 0, 1, 2, 3, 4 |

- When protocol is mysql:
  
| **Label** | **Example** | **Notes** |
| --- | --- | --- |
| request_content | select employee | SQL of MySQL. SQL has been truncated to avoid high-cardinality. The format is ['operation' 'space' 'table']. |
| response_content |  | Empty temporarily. |

- When protocol is kafka:
  
| **Label** | **Example** | **Notes** |
| --- | --- | --- |
| request_content | user-msg-topic | Topic of Kafka request. |
| response_content |  | Empty temporarily. |

- For other cases, the `request_content` and `response_content` are both empty.

## Topology Metrics

Topology metrics are typically generated from the client-side events, which are used to show the service dependencies map, so the metrics are called `topology`. Some timeseries may be generated from the server-side events, which contain a non-empty label `dst_container_id`. These metrics are useful when there is no agent installed on the client-side. 

| **Metric Name** | **Type** | **Description** |
| --- | --- | --- |
| kindling_topology_request_total | Counter | Total number of requests |
| kindling_topology_request_duration_nanoseconds_total | Counter |  Total duration of requests |
| kindling_topology_request_request_bytes_total | Counter | Total size of payload sent |
| kindling_topology_request_response_bytes_total | Counter | Total size of payload received |
| kindling_topology_request_average_duration_nanoseconds_count | Counter(Histogram) | Count of average duration of requests |​
| kindling_topology_request_average_duration_nanoseconds_sum | Counter(Histogram) | Sum of average duration of requests  |
| kindling_topology_request_average_duration_nanoseconds_bucket | Counter(Histogram) | Histogram buckets of average duration of requests |

| **Label Name** | **Example** | **Notes** |
| --- | --- | --- |
| src_node | slave-node1 | Which node the source pod is on |
| src_namespace | default | Namespace of the source pod |
| src_workload_kind | deployment | Workload kind of the source pod |
| src_workload_name | business1 | Workload name of the source pod |
| src_service | business1-svc | One of the services that target the source pod |
| src_pod | business1-0 | The name of the source pod |
| src_container | business-container | The name of the source container |
| src_container_id | 1a2b3c4d5e6f | The shorten container id which contains 12 characters |
| src_ip | 10.1.11.23 | The IP address of the source |
| dst_node | slave-node2 | Which node the destination pod is on |
| dst_namespace | default | Namespace of the destination pod |
| dst_workload_kind | deployment | Workload kind of the destination pod |
| dst_workload_name | business2 | Workload name of the destination pod |
| dst_service | business2-svc | One of the services that target the destination pod |
| dst_pod | business2-0 | The name of the destination pod |
| dst_container | business-container | The name of the source container |
| dst_container_id | 2b3c4d5e6f7e | (Only applicable to the timeseries generated from the server-side)The shorten container id which contains 12 characters |
| dst_ip | 10.1.11.24 | The IP address of the destination |
| dst_port | 80 | The listening port of the destination container  |
| protocol | http | The application layer protocol the requests use |
| status_code | 200 | Different values for different protocols  |

**The field "status_code" holds different values when "protocol" is different.**

- HTTP: 'Status Code' of HTTP response. 
- DNS: rcode of DNS response.
- others: empty temporarily

## Trace As Metric
We made some rules for considering whether a request is abnormal. For the abnormal request, the detail request information is considered as useful for debugging or profiling. We name this kind of data `trace`. It is not a good practice to store such data in Prometheus as some labels are high-cardinality, so we picked up some labels from the original ones to generate a new kind of metric, which is called `Trace As Metric`. The following table shows what labels this metric contains.  

| **Metric Name** | **Type** | **Description** |
| --- | --- | --- |
| kindling_trace_request_duration_nanoseconds | Gauge | The specific request duration |

| **Label Name** | **Example** | **Notes** |
| --- | --- | --- |
| src_node | slave-node1 | Which node the source pod is on |
| src_namespace | default | Namespace of the source pod |
| src_workload_kind | deployment | Workload kind of the source pod |
| src_workload_name | business1 | Workload name of the source pod |
| src_service | business1-svc | One of the services that target the source pod |
| src_pod | business1-0 | The name of the source pod |
| src_container | business-container | The name of the source container |
| src_container_id | 1a2b3c4d5e6f | (Only applicable when is_server is false)The shorten container id which contains 12 characters |
| src_ip | 10.1.11.23 | The IP address of the source |
| dst_node | slave-node2 | Which node the destination pod is on |
| dst_namespace | default | Namespace of the destination pod |
| dst_workload_kind | deployment | Workload kind of the destination pod |
| dst_workload_name | business2 | Workload name of the destination pod |
| dst_service | business2-svc | One of the services that target the destination pod |
| dst_pod | business2-0 | The name of the destination pod |
| dst_container | business-container | The name of the destination container |
| dst_container_id | 2b3c4d5e6f7e | (Only applicable when is_server is true)The shorten container id which contains 12 characters |
| dst_ip | 10.1.11.24 | The IP address of the destination |
| dst_port | 80 | The listening port of the destination container |
| protocol | http | The application layer protocol the requests use |
| is_server | true | True if the data is from the server-side, false otherwise |
| request_content | /test/api | Different values when protocol is different. Refer to service metric |
| response_content | 200 | Different values when protocol is different. Refer to service metric |
| request_duration_status | 1 | The total duration spent for sending request and receiving response.
1(green): latency <= 800ms
2(yellow): 800<latency<1500
3(red): latency >= 1500 |
| request_reqxfer_status | 2 |  ReqXfe indicates the duration for transferring request payload. 
1(green): latency <= 200ms
2(yellow): 200<latency<1000
3(red): latency >= 1000 |
| request_processing_status | 3 | Processing indicates the duration until receiving the first byte. 
1(green): latency <= 200ms
2(yellow): 200<latency<1000
3(red): latency >= 1000 |
| response_rspxfer_status | 1 | RspXfer indicates the duration for transferring response bopayloaddy.
1(green): latency <= 200ms
2(yellow): 200<latency<1000
3(red): latency >= 1000 |

## TCP (Layer 4) Metrics

| **Metric Name** | **Type** | **Description** |
| --- | --- | --- |
| kindling_tcp_rtt_microseconds | Gauge | Smoothed round trip time of the tcp socket |
| kindling_tcp_packet_loss_total | Counter | Total number of dropped packets |
| kindling_tcp_retransmit_total | Counter | Total times of retransmitting happens (not packets count) |

| **Label Name** | **Example** | **Notes** |
| --- | --- | --- |
| src_node | slave-node1 | Which node the source pod is on |
| src_namespace | default | Namespace of the source pod |
| src_workload_kind | deployment | Workload kind of the source pod |
| src_workload_name | business1 | Workload name of the source pod |
| src_service | business1-svc | One of the services that target the source pod |
| src_pod | business1-0 | The name of the source pod |
| src_container | business-container | The name of the source container |
| src_ip | 10.1.11.23 | Pod's IP by default. If the source is not a pod in kubernetes, this is the IP address of an external entity |
| src_port | 80 | The listening port of the source container, if applicable  |
| dst_node | slave-node2 | Which node the destination pod is on |
| dst_namespace | default | Namespace of the destination pod |
| dst_workload_kind | deployment | Workload kind of the destination pod |
| dst_workload_name | business2 | Workload name of the destination pod |
| dst_service | business2-svc | One of the services that target the destination pod |
| dst_pod | business2-0 | The name of the destination pod  |
| dst_container | business-container | The name of the destination container |
| dst_ip | 10.1.11.24 | Pod's IP by default. If the destination is not a pod in kubernetes, this is the IP address of an external entity |
| dst_port | 80 | The listening port of the destination container, if applicable |

## PromQL Example
Here are some examples of how to use these metrics in Prometheus, which can help you understand them faster.

| **Describe** | **PromQL** |
| --- | --- |
| Request counts | sum(increase(kindling_entity_request_duration_nanoseconds_count{namespace="$namespace",workload_name="$workload"}[$__interval])) by(namespace, workload_name) |
| DNS request counts | sum(increase(kindling_topology_request_duration_nanoseconds_count{src_namespace="$namespace",src_workload_name="$workload", protocol="dns"}[$__interval])) by (src_workload_name) |
| Latency | sum by(namespace, workload_name) (increase(kindling_entity_request_duration_nanoseconds_sum{namespace="$namespace", workload_name="$workload"}[$__interval])) / sum by(namespace, workload_name) (increase(kindling_entity_request_duration_nanoseconds_count{namespace="$namespace", workload_name="$workload"}[$__interval])) |
| Erroneous calls rate | sum (increase(kindling_entity_request_duration_nanoseconds_count{namespace="$namespace",workload_name="$workload",protocol="http",response_content=~"4..|5.."}[$__interval])) / sum (increase(kindling_entity_request_duration_nanoseconds_count{namespace="$namespace",workload_name="$workload",protocol="http"}[$__interval])) * 100 |
| Request latency quantile | histogram_quantile(0.99, rate(kindling_topology_request_duration_nanoseconds_bucket{dst_namespace="$namespace", dst_workload_name="$workload",protocol="http"}[$__interval])) and on (src_ip, src_pod, dst_ip, dst_pod)  kindling_trace_request_duration_nanoseconds{is_server="true", dst_namespace="$namespace", dst_workload_name="$workload", protocol="http"} |
| Retransmit times | sum(increase(kindling_tcp_retransmit_total{src_workload_name=~"$source", dst_workload_name=~"$destination"}[$__interval])) |
| Packets lost count | sum(increase(kindling_tcp_packet_loss_total{src_workload_name=~"$source", dst_workload_name=~"$destination"}[$__interval])) |
| Network sent bytes | sum(increase(kindling_topology_request_request_bytes_total{src_workload_name=~"$source", dst_workload_name=~"$destination"}[$__interval])) |
| Network received bytes | sum(increase(kindling_topology_request_response_bytes_total{src_workload_name=~"$source", dst_workload_name=~"$destination"}[$__interval])) |
