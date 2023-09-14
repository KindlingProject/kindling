import React, { useEffect, useState } from 'react';
import { Card } from '@grafana/ui';
import Flamegraph from './Flamegraph';
import Chart from './chart/chart';
import { cloneDeep, map } from 'lodash';

import CauseDataSource from '../causeDataSource';
import { flameDataHandle, basiclineOption, buildLineSeriesData } from '../service';


const testPQL = 'sum by (device) (rate(node_disk_io_time_seconds_total{instance=\"10.10.103.219:9100\"}[5m]))*100';
const causeDataQuery = new CauseDataSource();
interface Props {
    data: any;
}
function ChartStep({ data }: Props) {
    const [chartData, setChartData] = useState<any>();
    const [chartOption, setChartOption] = useState<any>(basiclineOption);
    
    const chartOptionHandle = (chartData: any) => {
        let option: any;
        if (data.chart === 'line') {
            option = cloneDeep(basiclineOption);
            option.xAxis[0].data = chartData[0].time;
            option.legend.data = map(chartData, 'name');
            chartData.forEach((item: any) => {
                option.series.push(buildLineSeriesData(item.name, item.data));
            });
            setChartOption(option);
        }
    }
    const dataHandle = () => {
        if (data.source === 'request' && data.request) {
            // 需要预先请求request提供的接口，请求对应绘制图表需要的数据
            causeDataQuery.excuteQuery(data.request).then(res => {
                if (data.chart === 'flame_graph') {
                    // 火焰图数据处理
                    const result = flameDataHandle(res.data);
                    setChartData(result);
                }
            })
        } else {
            // 数据已经直接返回的处理逻辑
            let chartData: any[] = [];
            Object.keys(data.metrics).forEach((key) => {
                if (data.metrics[key].result.length > 0) {
                    let cdItem = {
                        name: key,
                        action: data.metrics[key].action.text,
                        metric: data.metrics[key].result[0]?.metric,
                        time: data.metrics[key].result[0].values.map((item: any) => item[0] * 1000),
                        data: data.metrics[key].result[0].values.map((item: any) => item[1])
                    }
                    chartData.push(cdItem);
                }
            });
            chartData.length > 0 && chartOptionHandle(chartData);
        }   
    }

    useEffect(() => {
        dataHandle();
    // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [data]);

    const ChartComponent = () => {
        if (data.chart === 'flame_graph') {
            return <Flamegraph data={chartData}/>
        } else {
            return <Chart option={chartOption} height={260}/>
        }
    }

    return (
        <Card>
            <Card.Heading>{data.report_name}</Card.Heading>
            <Card.Description>
                {/* <TextTag data={data.text}/> */}
                {
                    ChartComponent()
                }
                {
                    data.conclusion ? <div>
                        <span>分析结论：</span>
                        <a className='conclusion_tag' href={`/d/c902c778-2a67-463c-8f08-ae495df7fd28/2?var-pql=${testPQL}&viewPanel=1`}>{data.conclusion}</a>
                    </div> : null
                }
            </Card.Description>
        </Card>
    );
}

export default ChartStep;
