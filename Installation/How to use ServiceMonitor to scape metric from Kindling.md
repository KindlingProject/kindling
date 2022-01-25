## Info


The Prometheus Operator (usually integrated into the **Kube Prometheus**) provides Kubernetes native deployment and management of Prometheus and related monitoring components.


It can automatically generate monitoring target configurations based on familiar Kubernetes label queries; no need to learn a Prometheus specific configuration language.


Follow the steps below,we can simplely add the Kindling-agent to the prometheus's targets as a data source.


## How to start


### 1. Confirm the environment of Prometheus-operator


In this turoial, we will use the Prometheus Operator . Use command below to confirm the installation of  Prometheus Operator.


_NAMESPACE、NAME  and VERSION may be different from example _


> $ kubectl get prometheus -A
> NAMESPACE    NAME   VERSION   REPLICAS   AGE
monitoring       k8s       v2.20.0       2                21d



### 2. Add a role to Prometheus's serviceAccount


Prometheus Operator needs permissions to watch the namespace where we installed Kindling-agent.
By Default , Prometheus Operator only have the permission to watch or list the pods in 'monitoring' namespace.


In this turoial,Kindling-agent is installed in the 'kindling' namespace.Here is a example to create a role and roleBindling for Prometheus,so that Prometheus can detect the serviceMonitor and kindling-agent we created.


> $ kubectl create -f kindling-rbac.yaml



**kindling-rbac.yaml:**


```yaml
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: prometheus-k8s
  namespace: kindling
rules:
- apiGroups:
  - ""
  resources:
  - services
  - endpoints
  - pods
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - extensions
  resources:
  - ingresses
  verbs:
  - get
  - list
  - watch

---  
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: prometheus-k8s
  namespace: kindling
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: prometheus-k8s
subjects:
- kind: ServiceAccount
  name: prometheus-k8s
  namespace: monitoring
```


### 3. Create a Service And ServiceMonitor for Kindling
ServiceMonitor is a CustomResources defined by Prometheus Operator.
​

The headless Service is used for collecting agents' podIp and port.
The ServiceMonitor is used for add the infomation to Prometheus targets which collected by headless Service.
> $ kubectl create -f kindling-service.yaml
> $ kubectl create -f kindling-serviceMonitor.yaml

**kindling-service.yaml:**


```yaml
---
apiVersion: v1
kind: Service
metadata:
  labels:
    k8s-app: kindling-agent
  name: kindling
  namespace: kindling
spec:
  clusterIP: None
  ports:
  - name: metrics
    port: 8080
    protocol: TCP
    targetPort: 8080
  selector:
    k8s-app: kindling-agent
  sessionAffinity: None
  type: ClusterIP
```


**kindling-servicemonitor.yaml:**


```yaml
---
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: kindling-k8s
  namespace: kindling
spec:
  endpoints:
  - interval: 15s
    port: metrics
    path: /metrics
    relabelings:
    - regex: '(container|endpoint|namespace|pod|service)'
      action: labeldrop
  namespaceSelector:
    matchNames:
    - kindling
  selector:
    matchLabels:
      k8s-app: kindling-agent
```
