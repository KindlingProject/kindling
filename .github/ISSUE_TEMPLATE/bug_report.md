---
name: Bug report
about: Create a report to help us improve
title: ''
labels: ''
assignees: ''

---

**Describe the bug**
A clear and concise description of what the bug is.
**What you expected to happen”
**Expected behavior**
A clear and concise description of what you expected to happen.
**How to reproduce**
Steps to reproduce the behavior.
**Screenshots**
If applicable, add screenshots to help explain your problem. 
**Logs**
Please attach the logs by running the following command:
```
kubectl logs -f kindling-agent-xxx(replace with your podname) -n kindling -c kindling-probe
kubectl logs -f kindling-agent-xxx(replace with your podname) -n kindling -c kindling-collector
```
**Environment (please complete the following information):**
- kindling agent version
- Node OS version
- K8s cluster version
- Node Kernel version
- Prometheus version
- Grafana version
**Additional context**
Add any other context about the problem here，like appliction protocol
