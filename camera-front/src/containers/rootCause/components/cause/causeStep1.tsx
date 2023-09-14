import React from 'react';
import { Descriptions, Tag, Collapse, Divider } from 'antd';
import TimeProgress from '../timeProgress';
import { formatUnit } from '@/services/util';


const Panel = Collapse.Panel;
const metric = 'kindling_span_trace_duration_nanoseconds_total';
interface IProps {
    data: any;
}
function CauseStep1({ data }: IProps) {
    return (
        <div className='step-card'>
            <div className='step-title'>故障节点分析</div>
            <div className='step-content'>
                <div>
                    <Descriptions>
                        <Descriptions.Item label="PodName">{ data.text.pod }</Descriptions.Item>
                        <Descriptions.Item label="请求URL">{ data.text.content_key }</Descriptions.Item>
                        <Descriptions.Item label="发生时间">{ formatUnit(data.timestamp / 1000000, 'date') }</Descriptions.Item>
                    </Descriptions>
                </div>
                {/* <TextTag data={data.text}/> */}
                <TimeProgress p90={data.metrics[metric] ? data.metrics[metric][0].baseline.value : 0} data={data.metrics[metric] ? data.metrics[metric][0].current.value : 0} />
                {
                    (data.logs && data.logs.length > 0) ? <React.Fragment>
                        <Divider plain>相关日志</Divider>
                        <Collapse accordion size="small">
                            {
                                data.logs.map(item => <Panel header={`线程名：${item.name}`} key={item.id}>
                                    <ul className='log_warp'>
                                        {
                                            item.values.map((opt, idx) => <li key={idx}>{opt}</li>)
                                        }
                                    </ul>
                                </Panel>)
                            }
                        </Collapse> 
                    </React.Fragment> : null
                }
                <div className='step-conclusion'>
                    <span>分析结论：</span>
                    <Tag color="error" className='custom-tag'>{data.conclusion}</Tag>
                </div>
            </div>
        </div>
    );
}

export default CauseStep1;
