import React, { useState } from 'react';
import { Card, Collapse } from '@grafana/ui';
import { formatUnit } from '../../../utils/utils.format';

const testPQL = 'sum by (device) (rate(node_disk_io_time_seconds_total{instance=\"10.10.103.219:9100\"}[5m]))*100';
const metric = 'kindling_span_trace_duration_nanoseconds_total';

interface Props {
    data: any;
}
function CauseStep1({ data }: Props) {
    const p90 = data.metrics[metric][0].baseline.value || 0;
    const error = data.metrics[metric][0].current.value || 0;
    const widthPercent = p90 / error * 80;

    const [openId, setOpenId] = useState<number | null>(null);

    const clickCollapse = (id: number) => {
        if (id === openId) {
            setOpenId(null);
        } else {
            setOpenId(id);
        }
    }

    return (
        <Card>
            <Card.Heading>故障节点分析</Card.Heading>
            <Card.Meta>{[`PodName: ${data.text.pod}`, `请求URL: ${data.text.content_key}`, `发生时间: ${formatUnit(data.timestamp / 1000000, 'date')}`]}</Card.Meta>
            <Card.Description>
                <div className='duration_warp'>
                    <div className='duration_item'>
                        <span>历史数据P90</span>
                        {
                            p90 > 0 ? <React.Fragment>
                                <div className='health_line' style={{ width: `${widthPercent}%` }}></div>
                                <span>{ formatUnit(p90, 'ns') }</span>
                            </React.Fragment> : null
                        }
                    </div>
                    <div className='duration_item'>
                        <span>本次故障调用</span>
                        {
                            error > 0 ? <React.Fragment>
                                <div className='abnormal_line'></div>
                                <span>{ formatUnit(error, 'ns') }</span>
                            </React.Fragment> : null
                        }
                    </div>
                </div>
                {
                    (data.logs && data.logs.length > 0) ? <React.Fragment>
                        <label style={{ marginBottom: '5px' }}>相关日志</label>
                        {
                            data.logs.map((item: any) => <Collapse collapsible label={`线程名：${item.name}`} key={item.id} isOpen={openId === item.id} onToggle={() => clickCollapse(item.id)}>
                                <ul className='log_warp'>
                                    {
                                        item.values.map((opt: any, idx: number) => <li key={idx}>{opt}</li>)
                                    }
                                </ul>
                            </Collapse>)
                        }
                    </React.Fragment> : null
                }
                <div>
                    <span>分析结论：</span>
                    <a className='conclusion_tag' href={`/d/c902c778-2a67-463c-8f08-ae495df7fd28/2?var-pql=${testPQL}&viewPanel=1`}>{data.conclusion}</a>
                </div>
            </Card.Description>
        </Card>
    );
}

export default CauseStep1;
