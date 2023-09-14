import React, { useEffect, useState } from 'react';
import { PanelProps } from '@grafana/data';
import { locationService } from '@grafana/runtime';

import { Card } from '@grafana/ui';
import CauseDataSource from './causeDataSource';
import './index.scss';

const causeDataQuery = new CauseDataSource();

interface Props extends PanelProps {}
function TracePanel(props: Props) {
    const { replaceVariables } = props;
    /**
     * grafana 组件props 中是有提供timeRange字段的，在开发自定义panel的时候可以监控timeRange变化
     * 但是在scene app的自定义plugin中不知道为什么timeRange一直没有改变，反正也是没有搜到解决方案
     * 最后使用grafana 默认全局提供的$__from和$__to两个变量获取当前选择对应的时间的值
     */
    const from = replaceVariables('$__from');
    const to = replaceVariables('$__to');
    const pid = replaceVariables('$pid');
    const url = replaceVariables('$url');
    const traceId = replaceVariables('$traceId');
    const [traceList, setTraceList] = useState([]);
    const [_location, setLocation] = useState(locationService.getLocation());

    useEffect(() => {
        const history = locationService.getHistory();
        const unlisten = history.listen((location: any) => {
            // console.log('location', location);
            setLocation(location)
        })
        return unlisten
    }, []);

    useEffect(() => {
        const params: any = {
            start: parseInt(from, 10) * 1000000,
            end: parseInt(to, 10) * 1000000
        }
        if (pid) {
            params.pid = parseInt(pid, 10);
        }
        if (url) {
            params.url = url;
        }
        if (traceId) {
            params.traceId = traceId;
        }
        causeDataQuery.excutePostQuery('/query/traceIds', params).then(res => {
            setTraceList(res.data);
        });
    }, [from, to, pid, url, traceId]);

    return (
        <div className='trace_list_warp'>
            {
                traceList.map((item: any) => <Card key={item.timestamp}>
                    <Card.Description>
                        <a href={`/a/kindling-app/cause/report/${item.traceId}`}>{item.traceId}</a>
                    </Card.Description>
                </Card>)
            }
        </div>
    );
}

export default TracePanel;

