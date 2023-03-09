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

export const query = async(queryText: string, timeRange: TimeRange) => {
    const kindlingDataSourceName = 'Prometheus';
    // const prometheusDataSource = await getDataSourceSrv().get(kindlingDataSourceName);
    // prometheusDataSource.query();
    const prometheusDataSource = getDataSourceSrv().getInstanceSettings(kindlingDataSourceName);
    // @ts-ignore
    let prometheusUrl = prometheusDataSource?.jsonData.directUrl;
    let api = `${prometheusUrl}/api/v1/query`;

    let timeScopeValue = '';
    if (_.isString(timeRange.raw.from) && _.isString(timeRange.raw.to)) {
        timeScopeValue = timeRange.raw.from.split('-')[1];
    } else {
        timeScopeValue = `${timeRange.to.diff(timeRange.from, 'm')}m`;
    }
    const nowTime = timeRange.to.toISOString();
    const timeScope: ScopedVars = {
        '__range': {
        text: '__range', 
        value: timeScopeValue
    }};
    let promql = getTemplateSrv().replace(queryText, timeScope);
    console.log(promql, nowTime);

    const response: any = await request(api, `query=${promql}&time=${nowTime}`);
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
export const metricQuery = async (metric: MetricType, timeRange: TimeRange) => {
    // console.log(metric, timeRange);
    let data: any;
    if (metric === 'sentVolume') {
        let promQL = 'increase(kindling_topology_request_request_bytes_total[$__range])';
        data = await query(promQL, timeRange);
    }
    if (metric === 'receiveVolume') {
        let promQL = 'increase(kindling_topology_request_response_bytes_total[$__range])';
        data = await query(promQL, timeRange);
    }
    if (metric === 'retransmit') {
        let promQL = 'increase(kindling_tcp_retransmit_total[$__range])';
        data = await query(promQL, timeRange);
    }
    if (metric === 'rtt') {
        let promQL = 'avg_over_time(kindling_tcp_srtt_microseconds[$__range])';
        data = await query(promQL, timeRange);
    }
    if (metric === 'packageLost') {
        let promQL = 'increase(kindling_tcp_packet_loss_total[$__range])';
        data = await query(promQL, timeRange);
    }
    if (metric === 'connFail') {
        let promQL = 'increase(kindling_tcp_connect_total[$__range])';
        data = await query(promQL, timeRange);
    }
    return data;
}

export const getNodeInfo = (timeRange: TimeRange) => {
    const callsPQL = 'sum (increase(kindling_entity_request_total[$__range])) by (namespace, workload_name, pod)';
    const timePQL = 'sum (increase(kindling_entity_request_duration_nanoseconds_total[$__range])) by (namespace, workload_name, pod)';
    const sendVolumePQL = 'sum(increase(kindling_entity_request_send_bytes_total[$__range])) by(namespace, workload_name, pod)';
    const receiveVolumePQL = 'sum(increase(kindling_entity_request_receive_bytes_total[$__range])) by(namespace, workload_name, pod)';
    return Promise.all([
        query(callsPQL, timeRange),
        query(timePQL, timeRange),
        query(sendVolumePQL, timeRange),
        query(receiveVolumePQL, timeRange)
    ]); 
}
