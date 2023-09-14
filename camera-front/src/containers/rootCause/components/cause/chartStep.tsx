import React, { useEffect, useState } from 'react';
import { Tag } from 'antd';
import { cloneDeep, map } from 'lodash';
import Flamegraph from './Flamegraph';
import Chart from './Chart';
import TextTag from './textTag';

import { getCauseReportsMoreInfo } from '@/request';
import { flameDataHandle } from '../../service';
import { basiclineOption, buildLineSeriesData } from '@/services/chartOption';

interface IProps {
    data: any;
}
function TableStep({ data }: IProps) {
    const [chartData, setChartData] = useState<any>();
    const [chartOption, setChartOption] = useState<any>(basiclineOption);
    
    const chartOptionHandle = (chartData) => {
        let option;
        if (data.chart === 'line') {
            option = cloneDeep(basiclineOption);
            option.xAxis[0].data = chartData[0].time;
            option.legend.data = map(chartData, 'name');
            chartData.forEach(item => {
                option.series.push(buildLineSeriesData(item.name, item.data));
            });
            // option.tooltip.formatter = (params) => formatChartTooltip(params, '%', total, '核');
            setChartOption(option);
        }
    }
    const dataHandle = () => {
        if (data.source === 'request' && data.request) {
            // 需要预先请求request提供的接口，请求对应绘制图表需要的数据
            getCauseReportsMoreInfo(data.request).then(res => {
                if (data.chart === 'flame_graph') {
                    // 火焰图数据处理
                    const result = flameDataHandle(res.data);
                    console.log(result);
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
                        time: data.metrics[key].result[0].values.map(item => item[0] * 1000),
                        data: data.metrics[key].result[0].values.map(item => item[1])
                    }
                    chartData.push(cdItem);
                }
            });
            chartData.length > 0 && chartOptionHandle(chartData);
        }   
    }

    useEffect(() => {
        dataHandle();
    }, [data]);

    const ChartComponent = () => {
        if (data.chart === 'flame_graph') {
            return <Flamegraph data={chartData}/>
        } else {
            return <Chart option={chartOption} height={260}/>
        }
    }

    return (
        <div className='step-card'>
            <div className='step-title'>{data.report_name}</div>
            <div className='step-content'>
                {/* <TextTag data={data.text}/> */}
                {
                    ChartComponent()
                }
                {
                    data.conclusion ? <div className='step-conclusion'>
                        <span>分析结论：</span>
                        <Tag color="error" className='custom-tag'>{data.conclusion}</Tag>
                    </div> : null
                }
            </div>
        </div>
    );
}

export default TableStep;
