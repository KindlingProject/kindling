import React, { useEffect, useState } from 'react';
import { Card, Table } from '@grafana/ui';
import { applyFieldOverrides, DataFrame, createTheme, FieldType, ArrayVector, Field, toDataFrame } from '@grafana/data';

const testPQL = 'sum by (device) (rate(node_disk_io_time_seconds_total{instance=\"10.10.103.219:9100\"}[5m]))*100';
const metric = "kindling_profiling_duration_nanoseconds_total";
interface Props {
    data: any;
}

function applyOverrides(dataFrame: DataFrame) {
    const dataFrames = applyFieldOverrides({
        data: [dataFrame],
        fieldConfig: {
            defaults: {},
            overrides: [],
        },
        replaceVariables: (value, vars, format) => {
            return vars && value === '${__value.text}' ? '${__value.text} interpolation' : value;
        },
        timeZone: 'utc',
        theme: createTheme(),
    });
    return dataFrames[0];
}

/**
 * 使用grafana ui提供的原生table组件，需要定义对应的DataFrame数据格式
 * 但是grafana table针对一列进行相关的配置定义，针对一列下单独某个数据进行自定义渲染的时候 有局限性。
 */
function CauseStep2({data}: Props) {
    const [eventList, setEventList] = useState<DataFrame>({fields: [], length: 0});

    useEffect(() => {
        const list: Field[] = [];
        const events = data.metrics[metric] || [];
        console.log('events', events);
        list.push({
            name: 'cpuType',
            type: FieldType.string,
            values: new ArrayVector(events.map((opt: any) => opt.labels.cpuType)),
            config: {}
        });
        list.push({
            name: 'p90',
            type: FieldType.number,
            values: new ArrayVector(events.map((opt: any) => opt.baseline.value)),
            config: {
                unit: events[0].unit
            }
        });
        list.push({
            name: 'current',
            type: FieldType.number,
            values: new ArrayVector(events.map((opt: any) => opt.current.value)),
            config: {
                unit: events[0].unit
            }
        });
        console.log('list', list);
        const dataFrame = toDataFrame({
            fields: list,
            length: events.length
        });
        setEventList(applyOverrides(dataFrame));
    }, [data])

    return (
        <Card>
            <Card.Heading>接口执行耗时分析</Card.Heading>
            <Card.Description>
                <Table data={eventList} width={800} height={300} />
                <div>
                    <span>分析结论：</span>
                    <a className='conclusion_tag' href={`/d/c902c778-2a67-463c-8f08-ae495df7fd28/2?var-pql=${testPQL}&viewPanel=1`}>{data.conclusion}</a>
                </div>
            </Card.Description>
        </Card>
    );
}

export default CauseStep2;
