import { getDataSourceSrv, getBackendSrv, getTemplateSrv } from '@grafana/runtime';
import { TimeRange, ScopedVars } from '@grafana/data';
import { MetricType, DataSourceResponse } from './types';
import { lastValueFrom } from 'rxjs';
import _ from 'lodash';

const request = async(url: string, params?: string) => {
    const response: any = getBackendSrv().fetch<DataSourceResponse>({
      url: `${url}${params?.length ? `?${params}` : ''}`,
    });
    return lastValueFrom(response);
}
// 获取下拉列表中的namespace和workload label值
export const getNamespaceAndWorkload = async (timeRange: TimeRange) => {
    const kindlingDataSourceName = 'Prometheus';
    const prometheusDataSource = getDataSourceSrv().getInstanceSettings(kindlingDataSourceName);
    // @ts-ignore
    let prometheusUrl = prometheusDataSource?.jsonData.directUrl;
    let api = `${prometheusUrl}/api/v1/series`;

    const startTime = timeRange.from.toISOString();
    const endTime = timeRange.to.toISOString();

    return request(api, `match[]=${'kindling_topology_request_total'}&start=${startTime}&end=${endTime}`);
}
// Prometheus api接口返回的数据格式进行转化
const transformResponseData = (data: any) => {
    let result: any[] = [];
    _.forEach(data, item => {
        let tdata: any = {
            ...item.metric,
            values: [item.value[1]]
        }
        result.push(tdata);
    });
    return result;
}
// 获取timeRange 的时间range值
const getTimeScope = (timeRange: TimeRange) => {
    let timeScopeValue = '';
    if (_.isString(timeRange.raw.from) && _.isString(timeRange.raw.to)) {
        timeScopeValue = timeRange.raw.from.split('-')[1];
    } else {
        timeScopeValue = `${timeRange.to.diff(timeRange.from, 'm')}m`;
    }
    const timeScope: ScopedVars = {
        '__range': {
        text: '__range', 
        value: timeScopeValue
    }};

    return timeScope;
}

const excuteQuery = async(queryText: string, timeRange: TimeRange) => {
    const kindlingDataSourceName = 'Prometheus';
    // const prometheusDataSource = await getDataSourceSrv().get(kindlingDataSourceName);
    // prometheusDataSource.query();
    const prometheusDataSource = getDataSourceSrv().getInstanceSettings(kindlingDataSourceName);
    // @ts-ignore
    let prometheusUrl = prometheusDataSource?.jsonData.directUrl;
    let api = `${prometheusUrl}/api/v1/query`;

    const nowTime = timeRange.to.toISOString();
    const response: any = await request(api, `query=${queryText}&time=${nowTime}`);
    if (response.data.status) {
        const result = transformResponseData(response.data.data.result);
        return Promise.resolve(result);
    } else {
        console.error(response.statusText);
        return Promise.resolve([]);
    }
}

/**
 * 切换指标时再查询对应的指标详细信息数据
 * @param metric 
 * @param timeRange 
 */
export const metricQuery = async (metric: MetricType, namespace: string, workload: string, timeRange: TimeRange) => {
    // console.log(metric, timeRange);
    let data: any;
    if (metric === 'sentVolume') {
        // let promQL = 'increase(kindling_topology_request_request_bytes_total[$__range])';
        let promQL = getQueryText('kindling_topology_request_request_bytes_total', namespace, workload, timeRange);
        data = await excuteQuery(promQL, timeRange);
    }
    if (metric === 'receiveVolume') {
        // let promQL = 'increase(kindling_topology_request_response_bytes_total[$__range])';
        let promQL = getQueryText('kindling_topology_request_response_bytes_total', namespace, workload, timeRange);
        data = await excuteQuery(promQL, timeRange);
    }
    if (metric === 'retransmit') {
        // let promQL = 'increase(kindling_tcp_retransmit_total[$__range])';
        let promQL = getQueryText('kindling_tcp_retransmit_total', namespace, workload, timeRange);
        data = await excuteQuery(promQL, timeRange);
    }
    if (metric === 'rtt') {
        // let promQL = 'avg_over_time(kindling_tcp_srtt_microseconds[$__range])';
        let promQL = getQueryText('kindling_tcp_srtt_microseconds', namespace, workload, timeRange);
        data = await excuteQuery(promQL, timeRange);
    }
    if (metric === 'packageLost') {
        // let promQL = 'increase(kindling_tcp_packet_loss_total[$__range])';
        let promQL = getQueryText('kindling_tcp_packet_loss_total', namespace, workload, timeRange);
        data = await excuteQuery(promQL, timeRange);
    }
    if (metric === 'connFail') {
        // let promQL = 'increase(kindling_tcp_connect_total[$__range])';
        let promQL = getQueryText('kindling_tcp_connect_total', namespace, workload, timeRange);
        data = await excuteQuery(promQL, timeRange);
    }
    return data;
}


const getNodeMetricPQL = (metric: string, namespace: string, workload: string, timeRange: TimeRange) => {
    const timeScope: ScopedVars = getTimeScope(timeRange);
    let queryPQL = '';
    if (namespace === 'all') {
        queryPQL = `sum (increase(${metric}[$__range])) by (namespace, workload_name, pod)`;
    } else {
        queryPQL = `sum (increase(${metric}{namespace="${namespace}"}[$__range])) by (namespace, workload_name, pod)`;
        // if (workload === 'all') {
        //     queryPQL = `sum (increase(${metric}{namespace="${namespace}"}[$__range])) by (namespace, workload_name, pod)`;
        // } else {
        //     queryPQL = `sum (increase(${metric}{namespace="${namespace}", workload_name="${workload}"}[$__range])) by (namespace, workload_name, pod)`;
        // }
    }
    let promql = getTemplateSrv().replace(queryPQL, timeScope);
    return promql;
}
/**
 * 获取节点上的相关指标数据
 * @param timeRange 
 * @returns 
 */
export const getNodeInfo = (namespace: string, workload: string, timeRange: TimeRange) => {
    // const callsPQL = 'sum (increase(kindling_entity_request_total[$__range])) by (namespace, workload_name, pod)';
    // const timePQL = 'sum (increase(kindling_entity_request_duration_nanoseconds_total[$__range])) by (namespace, workload_name, pod)';
    // const sendVolumePQL = 'sum(increase(kindling_entity_request_send_bytes_total[$__range])) by(namespace, workload_name, pod)';
    // const receiveVolumePQL = 'sum(increase(kindling_entity_request_receive_bytes_total[$__range])) by(namespace, workload_name, pod)';
    const callsPQL = getNodeMetricPQL('kindling_entity_request_total', namespace, workload, timeRange);
    const timePQL = getNodeMetricPQL('kindling_entity_request_duration_nanoseconds_total', namespace, workload, timeRange);
    const sendVolumePQL = getNodeMetricPQL('kindling_entity_request_send_bytes_total', namespace, workload, timeRange);
    const receiveVolumePQL = getNodeMetricPQL('kindling_entity_request_receive_bytes_total', namespace, workload, timeRange);
    return Promise.all([
        excuteQuery(callsPQL, timeRange),
        excuteQuery(timePQL, timeRange),
        excuteQuery(sendVolumePQL, timeRange),
        excuteQuery(receiveVolumePQL, timeRange)
    ]); 
}

const getQueryText = (metric: string, namespace: string, workload: string, timeRange: TimeRange) => {
    const timeScope: ScopedVars = getTimeScope(timeRange);
    let queryPQL = '';
    let functionName = 'increase';
    if (metric === 'kindling_tcp_srtt_microseconds') {
        functionName = 'avg_over_time';
    }
    if (namespace === 'all') {
        queryPQL = `${functionName}(${metric}[$__range])`;
    } else {
        queryPQL = `${functionName}(${metric}{src_namespace="${namespace}"}[$__range]) or ${functionName}(${metric}{dst_namespace="${namespace}"}[$__range])`;
        // if (workload === 'all') {
        //     queryPQL = `${functionName}(${metric}{src_namespace="${namespace}"}[$__range]) or ${functionName}(${metric}{dst_namespace="${namespace}"}[$__range])`;
        // } else {
        //     queryPQL = `${functionName}(${metric}{src_namespace="${namespace}", src_workload_name="${workload}"}[$__range]) or ${functionName}(${metric}{dst_namespace="${namespace}", src_workload_name="${workload}"}[$__range])`;
        // }
    }
    let promql = getTemplateSrv().replace(queryPQL, timeScope);
    return promql;
}
// 初始化请求拓扑数据
export const getTopoData = (namespace: string, workload: string, timeRange: TimeRange) => {
    let promQL1 = getQueryText('kindling_topology_request_total', namespace, workload, timeRange);
    let promQL2 = getQueryText('kindling_topology_request_duration_nanoseconds_total', namespace, workload, timeRange);
    console.log(promQL1, promQL2);
    return Promise.all([
        excuteQuery(promQL1, timeRange),
        excuteQuery(promQL2, timeRange)
    ]); 
}
