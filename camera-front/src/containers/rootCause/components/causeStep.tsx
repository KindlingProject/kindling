import React from 'react';
import { ArrowDownOutlined } from '@ant-design/icons'; 
import CauseStep1 from './cause/causeStep1';
import CauseStep2 from './cause/causeStep2';
import TableStep from './cause/tableStep';
import ChartStep from './cause/chartStep';

interface IProps {
    data: any;
    showArrow: boolean;
}

function CauseStep({ data, showArrow }: IProps) {

    const handleStep = () => {
        if (data.type === 'trace_entry') {
            return <CauseStep1 data={data}/>
        } else if (data.type === 'profiling_duration') {
            return <CauseStep2 data={data}/>
        } else if (data.type === 'metrics_baseline' || data.type === 'tableMetrics') {
            return <TableStep data={data}/>
        } else if (data.type === 'metrics_timeseries') {
            return <ChartStep data={data}/>
        } else if (data.type === 'display') {
            return <ChartStep data={data}/>
        } else {
            return <div>暂未匹配类型</div>
        }
    }

    return (
        <div className='step_warp'>
            {handleStep()}
            {
                showArrow ? <ArrowDownOutlined /> : null
            }
        </div>
    )
    
}

export default CauseStep;
