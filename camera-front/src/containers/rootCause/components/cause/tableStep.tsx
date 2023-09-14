import React, { useEffect, useState } from 'react';
import { Tag, Table, Popover } from 'antd';
import TextTag from './textTag';
import { formatUnit } from '@/services/util';

const staticColumns = [
    {
        title: '指标',
        dataIndex: 'name'
    }, {
        title: '标签',
        dataIndex: 'labels',
        width: '50%',
        render: (_text, record: ITData) => {
            return <div>
                {
                    Object.keys(record.labels).map((key, idx) => <Tag key={idx}>{`${key}:${record.labels[key]}`}</Tag>)
                }
            </div>
        }
    }
]
interface ITData {
    name: string;
    baseline: number;
    baseAction: string;
    current: number;
    currentAction: string;
    unit: any;
    warn: boolean;
    labels: {
        [key in string]: any
    }
}
interface IProps {
    data: any;
}
function TableStep({ data }: IProps) {
    const [columns, setColumns] = useState([]);
    const [tableData, setTableData] = useState<ITData[]>([]);

    const dataHandle = () => {
        const list: ITData[] = [];
        let metrics = Object.keys(data.metrics);
        const columnList: any = [...staticColumns];
        if (data.type === 'metrics_baseline') {
            columnList.splice(1, 0, ...[{
                title: '历史基线',
                dataIndex: 'baseline',
                render: (text, record) => {
                    if (record.baseAction) {
                        return <Popover content={record.baseAction}>
                            <span>{record.unit ? formatUnit(text, record.unit) : text}</span>
                        </Popover>
                    } else {
                        return <span>{record.unit ? formatUnit(text, record.unit) : text}</span>
                    }
                }
            }, {
                title: '故障调用',
                dataIndex: 'current',
                render: (text, record: ITData) => {
                    if (record.currentAction) {
                        return <Popover content={record.currentAction}>
                            <span className={record.warn ? 'a_error' : ''}>{record.unit ? formatUnit(text, record.unit) : text}</span>
                        </Popover>
                    } else {
                        return <span className={record.warn ? 'a_error' : ''}>{record.unit ? formatUnit(text, record.unit) : text}</span>
                    }
                }
            }])
        } else {
            columnList.splice(1, 0, {
                title: '值',
                dataIndex: 'current',
                render: (text, record: ITData) => {
                    if (record.currentAction) {
                        return <Popover content={record.currentAction}>
                            <span className={record.warn ? 'a_error' : ''}>{record.unit ? formatUnit(text, record.unit) : text}</span>
                        </Popover>
                    } else {
                        return <span className={record.warn ? 'a_error' : ''}>{record.unit ? formatUnit(text, record.unit) : text}</span>
                    }
                }
            });
        }
        metrics.forEach(metric => {
            data.metrics[metric].forEach(item => {
                list.push({
                    name: metric,
                    baseline: item.baseline?.value,
                    baseAction: item.baseline?.action?.text || '',
                    current: item.current.value,
                    currentAction: item.current?.action?.text || '',
                    unit: item.unit,
                    warn: item.is_warn,
                    labels: item.labels
                })
            });
        });
        console.log(list);
        setColumns(columnList);
        setTableData(list);
    }

    useEffect(() => {
        dataHandle();
    }, [data]);

    return (
        <div className='step-card'>
            <div className='step-title'>{data.report_name}</div>
            <div className='step-content'>
                <TextTag data={data.text}/>
                <Table size='small' columns={columns} dataSource={tableData} bordered pagination={false}></Table>
                <div className='step-conclusion'>
                    <span>分析结论：</span>
                    <Tag color="error" className='custom-tag'>{data.conclusion}</Tag>
                </div>
            </div>
        </div>
    );
}

export default TableStep;
