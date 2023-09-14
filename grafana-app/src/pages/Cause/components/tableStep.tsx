import React, { useEffect, useState } from 'react';
import { Card, Badge, Tooltip } from '@grafana/ui';
import DataTable from './dataTable';
import { formatUnit } from '../../../utils/utils.format';

const staticColumns = [
    {
        title: '指标',
        accessor: 'name'
    }, {
        title: '标签',
        accessor: 'labels',
        Cell: (cell: any) => {
            const { original } = cell.cell.row;
            return <div>
                {
                    Object.keys(original.labels).map((key, idx) => <Badge key={idx} color='blue' text={`${key}:${original.labels[key]}`}/>)
                }
            </div>
        }
    }
];

interface TData {
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
interface Props {
    data: any;
}
function TableStep({ data }: Props) {
    const [columns, setColumns] = useState([]);
    const [tableData, setTableData] = useState<TData[]>([]);

    const dataHandle = () => {
        const list: TData[] = [];
        let metrics = Object.keys(data.metrics);
        const columnList: any = [...staticColumns];
        if (data.type === 'metrics_baseline') {
            columnList.splice(1, 0, ...[{
                title: '历史基线',
                accessor: 'baseline',
                Cell: (cell: any) => {
                    const { value: text } = cell;
                    const { original: record } = cell.cell.row;
                    if (record.baseAction) {
                        return <Tooltip content={record.baseAction}>
                            <span>{record.unit ? formatUnit(text, record.unit) : text}</span>
                        </Tooltip>
                    } else {
                        return <span>{record.unit ? formatUnit(text, record.unit) : text}</span>
                    }
                }
            }, {
                title: '故障调用',
                accessor: 'current',
                Cell: (cell: any) => {
                    const { value: text } = cell;
                    const { original: record } = cell.cell.row;
                    if (record.currentAction) {
                        return <Tooltip content={record.currentAction}>
                            <span className={record.warn ? 'a_error' : ''}>{record.unit ? formatUnit(text, record.unit) : text}</span>
                        </Tooltip>
                    } else {
                        return <span className={record.warn ? 'a_error' : ''}>{record.unit ? formatUnit(text, record.unit) : text}</span>
                    }
                }
            }])
        } else {
            columnList.splice(1, 0, {
                title: '值',
                accessor: 'current',
                Cell: (cell: any) => {
                    const { value: text } = cell;
                    const { original: record } = cell.cell.row;
                    if (record.currentAction) {
                        return <Tooltip content={record.currentAction}>
                            <span className={record.warn ? 'a_error' : ''}>{record.unit ? formatUnit(text, record.unit) : text}</span>
                        </Tooltip>
                    } else {
                        return <span className={record.warn ? 'a_error' : ''}>{record.unit ? formatUnit(text, record.unit) : text}</span>
                    }
                }
            });
        }
        metrics.forEach(metric => {
            data.metrics[metric].forEach((item: any) => {
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
        console.log(columnList, list);
        setColumns(columnList);
        setTableData(list);
    }

    useEffect(() => {
        dataHandle();
    // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [data]);

    return (
        <Card>
            <Card.Heading>{data.report_name}</Card.Heading>
            <Card.Description>
                <DataTable columns={columns} data={tableData}/>
                <div>
                    <span>分析结论：</span>
                    <a className='conclusion_tag'>{data.conclusion}</a>
                </div>
            </Card.Description>
        </Card>
    );
}

export default TableStep;
