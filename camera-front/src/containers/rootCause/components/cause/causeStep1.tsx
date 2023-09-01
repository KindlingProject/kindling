import React from 'react';
import { Descriptions, Tag } from 'antd';
import TimeProgress from '../timeProgress';
import TextTag from './textTag';
import { formatUnit } from '@/services/util';


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
                <div className='step-conclusion'>
                    <span>分析结论：</span>
                    <Tag color="error">{data.conclusion}</Tag>
                </div>
            </div>
        </div>
    );
}

export default CauseStep1;
