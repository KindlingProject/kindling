# Kindling

# [![](https://img.shields.io/badge/license-Apache2.0-blue.svg)](https://github.com/Kindling-project/kindling/blob/main/LICENSE)[![](https://img.shields.io/badge/docs-latest-green.svg)](http://www.kindling.space:33215/project-1/)

## What is Kindling

Kindling is an eBPF-based cloud-native monitor tool, which aims to help users understand the app behavior from kernel to code stack. Currently, it provides an easy way to get the view of the network flows in the Kubernetes environment, and many built-in network monitor dashboards like retransmit, DNS, throughput, TPS. Not only as a network monitor, but Kindling is also trying to analyze one detail RPC call and get the following information, which can be found in network explore in chrome. And the usage is the same as network explore in chrome, with which user can tell which part to dig into to find the root cause of the problem in the production environment. 

![img](https://cdn.nlark.com/yuque/0/2021/png/749988/1633767195234-af2913c4-75d7-447d-99bb-bd1d815883ae.png)

# Architecture

From a high-level view, the agent runs as DeamonSet in Kuberntes. It collects all the SYSCALL and some other tracepoints. We use different exporters for different distributions. For the lightweight version, we just build a Prometheus exporter to export the data which can be stored into Prometheus and displayed in Grafana Plugin. For the standard version, which is designed for heavy usage, Kafka is adopted for buffering the events, and data is stored into ElasticSearch with much more detailed information which can be displayed. Currently, the light version has been open-sourced. 

![](https://cdn.nlark.com/yuque/0/2021/png/749988/1636076970508-2aa15679-2ab7-4f82-bd58-2ce3832b4550.png)

## Linux kernel version support

Kindling eBPF module depends on the kernel version which is larger than 4.14. As an eBPF constraint,  the eBPF module can't work for older kernel version. But for the users who want to try the functionality with the old kernel, we use some module from Sysdig open-source project with enhancement and verification. Basically, the idea is to use a kernel module for tracking the kernel trace-point. Thanks to Sysdig open-source project, it provides a trace-point instrument framework for older kernel version. 

For now, the kernel module works as expected as eBPF module during our test, but it is recommended to use eBPF module for production environment because it is safer than kernel module. In order to help old kernel version user experience the eBPF magic, we will support the kernel model. And you are welcome to report the issue for the kernel module. For the functionality, the kernel module and eBPF module capture the same data, and behave exactly the same.   

## Why we build Kindling?

When we talk about observability, we already have plenty of tools to use, like Skywalking for tracing, ELK for logging, Prometheus for metric. Why should we build an eBPF-based monitor tool?

The majority issue for user adoption of k8s is the complexity. For the apps on the kubernetes, we don't know the network flows between the service until we can instrument the apps. We can't tell which part to blame when there is a product issue arise. Do we configure the kubernetes correctly? Are there any bugs in the virtual network like Calico/Flannel that caused this problem? Does the app code cause this issue?

We are a company based in Hangzhou, China and used to provide the kubernetes distribution for our customers. Our customers used to have those questions, and we don't have a proper solution to answer those questions.

APM is good for those java language apps which can be instrumented automatically, but the code has to be rewritten for instrumenting the go apps. And even we adopt the APM solution, we still can't tell whether an issue is caused by network problems.

We found it may be helpful that we check the issue from the network view to identify issues roughly like "oh, it's a network problem, the code works fine, we should dig into the configuration of calico" or "the infrastructure works fine, the app code should be blamed, let's dig into the logs or the APM dashboard for further information" 

### Why eBPF

Libpcap way of analyzing the flows in the kubernetes environment is too expensive for CPU and network. eBPF way of data capture cost much less than libpcap. eBPF is the most popular technology to track the Linux kernel. And the virtual network is built by veth-pair and iptables, which works in the kernel. So eBPF should be used for tracking how the kernel respond to the app request.

## Core Features

![img](https://cdn.nlark.com/yuque/0/2022/png/749988/1642572876088-c26396ac-e7bb-44e7-ae0c-cc96f3344cd8.png)

Kindling provides two versions which has different feature but with the same agent. The lightweight is integrated into Prometheus, and it uses PromQL to query the data from Prometheus, so it should be adopted easily. But due to the cardinality constraint of Prometheus, we group the detailed data into buckets which throw away the detailed information. For the standard version, we provide much more detailed information because we use ElasticSearch as the back-end to store the original information. 

The lightweight version will be open sourced first and the standard version is under ative devlopment and will open sourced in next few months. 

| Feature                                          | Lightweight | Standard |
| ------------------------------------------------ | ----------- | -------- |
| Service Map                                      | √           | √        |
| DNS Monitor                                      | √           | √        |
| TCP Status Monitor                               | √           | √        |
| HTTP/MYSQL/REDIS detailed information            | √           | √        |
| Grafana plugin display                           | √           | √        |
| Network traffic dump                             |             | √        |
| Virtual network packet  tracing                  |             | √        |
| Kuberntes infrastructure monitor and integration |             | √        |
| Warning and alert                                |             | √        |
| Multi-cluster management                         |             | √        |

## Get started

You can deploy Kindling easily, check out [Installation Guide]() for detail.

## Documentation

The Kindling documentation is available at [kindling website]( http://www.kindling.space:33215/project-1/)

Our Roadmap

## Contributing 

Contributions are welcome, you can contribute in many ways: report issues,help us reproduce issues, fix bugs, add features, or give us advice on github discussions and so on.If you’re interested in joining us to unveil the e BPF  in kubernetes area,you can start by reading the [Contributing Guide](https://github.com/Kindling-project/kindling/blob/main/CONTRIBUTING.md).

## Contact

if  you have questions or ideas, feel free to reach out to us in the following ways:

- Check out our [discussions](https://github.com/Kindling-project/kindling/discussions)
- Join us on our [Slack team](https://kindling-world.slack.com/join/shared_invite/zt-12bdv5kbn-HPajCjB0_CmuvXw~4t_lAw#/shared-invite/email)
-  Join us from Wechat

![image-20220125192518485](/Users/jiang/Library/Application Support/typora-user-images/image-20220125192518485.png)

## License

Kindling is distributed under [Apache License, Version2.0](https://github.com/Kindling-project/kindling/blob/main/LICENSE).

