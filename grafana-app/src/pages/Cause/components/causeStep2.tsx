import React, { useEffect, useState } from 'react';
import { Card } from '@grafana/ui';
import DataTable from './dataTable';
import { formatUnit } from '../../../utils/utils.format';

const testPQL = 'sum by (device) (rate(node_disk_io_time_seconds_total{instance=\"10.10.103.219:9100\"}[5m]))*100';
const metric = "kindling_profiling_duration_nanoseconds_total";
interface EventProps {
    name: string;
    baseline: number;
    basePercent?: number;
    current: number;
    currentPercent?: number;
    warn: boolean;
    hide?: boolean;
}
interface Props {
    data: any;
}
function CauseStep2({data}: Props) {
    const [eventList, setEventList] = useState<EventProps[]>([]);

    useEffect(() => {
        const list: EventProps[] = [];
        const events = data.metrics[metric] || [];
        events.forEach((item: any) => {
            list.push({
                name: item.labels.cpuType,
                baseline: item.baseline.value,
                current: item.current.value,
                warn: item.is_warn
            });
        });
        setEventList(list);
    }, [data])

    const columns = [
        {
            title: '', 
            accessor: 'name',
            Cell: ({value}: any) => {
                return <a href='/d/node-cpu-test/cpu' target='_blank'>{value}</a>
            }
        }, 
        {
            title: '历史数据P90', 
            accessor: 'baseline',
            Cell: ({ value }: any) =>  formatUnit(value, 'ns')
        },
        {
            title: '本次调用数据', 
            accessor: 'current',
            Cell: (cell: any) =>  {
                const { value } = cell;
                const { original } = cell.cell.row;
                // console.log('props', cell, original);
                return <span style={{ color: original.warn ? '#ff3c3c' : '' }}>{formatUnit(value, 'ns')}</span>
            }
        }
    ]
    return (
        <Card>
            <Card.Heading>接口执行耗时分析</Card.Heading>
            <Card.Description>
                {/* <Table data={eventList} width={800} height={300} /> */}
                <DataTable columns={columns} data={eventList}/>
                <div>
                    <span>分析结论：</span>
                    <a className='conclusion_tag' href={`/d/c902c778-2a67-463c-8f08-ae495df7fd28/2?var-pql=${testPQL}&viewPanel=1`}>{data.conclusion}</a>
                </div>
            </Card.Description>
        </Card>
    );
}

export default CauseStep2;
