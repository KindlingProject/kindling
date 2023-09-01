import React, { useEffect, useState } from 'react';
import { Table, Tag } from 'antd'; 
import { chain, map } from 'lodash';
import { formatUnit } from '@/services/util';
import TextTag from './textTag';

const metric = "kindling_profiling_duration_nanoseconds_total";
const eventColors = [
    {
        name: 'cpu',
        color: '#FDE6E4'
    }, {
        name: 'file',
        color: '#EDE4FF'
    }, {
        name: 'net',
        color: '#FEF3E6'
    }, {
        name: 'futex',
        color: '#E6F4F5'
    }, {
        name: 'idle',
        color: '#EDF8FF'
    }, {
        name: 'epoll',
        color: '#e5f5ea'
    }, {
        name: 'other',
        color: '#E5E5E5'
    }
];
interface IEvent {
    name: string;
    baseline: number;
    basePercent?: number;
    current: number;
    currentPercent?: number;
    color: string;
    warn: boolean;
    hide?: boolean;
}
interface IProps {
    data: any;
}
function CauseStep2({data}: IProps) {
    const [eventList, setEventList] = useState<IEvent[]>([]);

    const dataHandle = () => {
        const list: IEvent[] = [];
        const events = data.metrics[metric] || [];
        events.forEach(item => {
            let color = eventColors.find(opt => opt.name === item.labels.cpuType)?.color || '#ececec';
            list.push({
                name: item.labels.cpuType,
                baseline: item.baseline.value,
                current: item.current.value,
                color,
                warn: item.is_warn
            });
        });
        setEventList(list);
    }

    useEffect(() => {
        dataHandle();
    }, [data])

    return (
        <div className='step-card'>
            <div className='step-title'>接口执行耗时分析</div>
            <TextTag data={data.text}/>
            <Table size='small' dataSource={eventList} rowKey="name" pagination={false} bordered>
                <Table.Column title="" dataIndex="name"></Table.Column>
                <Table.Column title="历史数据P90" dataIndex="baseline" render={(v) => formatUnit(v, 'ns')}></Table.Column>
                <Table.Column title="本次调用数据" dataIndex="current" render={(v, record: IEvent) =>  <span className={record.warn ? 'a_error' : ''}>{formatUnit(v, 'ns')}</span>}></Table.Column>
            </Table>
            <div className='step-conclusion'>
                <span>分析结论：</span>
                <Tag color="error">{data.conclusion}</Tag>
            </div>
        </div>
    );
}

export default CauseStep2;
