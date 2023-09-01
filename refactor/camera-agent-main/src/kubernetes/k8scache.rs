use serde::Serialize;

#[derive(Default, Debug, Serialize, Clone)]
pub struct K8sPodInfo {
    pub container_name: String,
    pub pod_ip: String,
    pub pod_name: String,
    pub workload_kind: String,
    pub workload_name: String,
    pub namespace: String,
}

pub struct K8sMetaDataCache {}

impl K8sMetaDataCache {
    pub fn new() -> K8sMetaDataCache {
        Self {
            // TODO Fix Me
        }
    }

    pub fn get_k8s_pod_info(&self, node_ip: &str, container_id: &str) -> K8sPodInfo {
        // TODO user node_ip & container_id query pod info.
        K8sPodInfo {
            container_name: "".to_string(),
            pod_ip: "".to_string(),
            pod_name: "".to_string(),
            workload_kind: "".to_string(),
            workload_name: "".to_string(),
            namespace: "".to_string(),
        }
    }
}
