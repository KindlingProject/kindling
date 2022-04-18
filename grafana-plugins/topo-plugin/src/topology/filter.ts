import _ from 'lodash';
export interface SelectOption {
    label: string, 
    value: string,
}

class FilterList {
    namespaceList: SelectOption[];
    workloadList: SelectOption[];
    workloadListByNamespace: any;

    constructor(data: any) {
        this.namespaceList = [];
        this.workloadList = [];
        this.workloadListByNamespace = {};

        this.init(data);
    }

    init(data: any[]) {
        let namespaces = _.uniq(_.concat(_.map(data, 'src_namespace'), _.map(data, 'dst_namespace')));
        this.namespaceList.push({
            label: 'All', 
            value: 'all'
        });
        this.workloadList.push({
            label: 'All', 
            value: 'all'
        });
        
        _.forEach(namespaces, namespace => {
            this.namespaceList.push({
                label: namespace, 
                value: namespace
            });

            let workloads = _.uniq(_.map(data, item => {
                if (item['src_namespace'] === namespace) {
                    return item['src_workload_name']
                }
                if (item['dst_namespace'] === namespace) {
                    return item['dst_workload_name']
                }
            }));
            this.workloadListByNamespace[namespace] = [];
            _.forEach(workloads, workload => {
                if (workload) {
                    this.workloadListByNamespace[namespace].push({
                        label: workload, 
                        value: workload
                    });
                }
            });
        });
    }

    changeNamespace(namespace: string) {
        let workloads = this.workloadListByNamespace[namespace];
        this.workloadList = _.concat([{
            label: 'All', 
            value: 'all'
        }], workloads);
    }
}
    
export default FilterList;
