import React from 'react';
import CauseStep1 from './causeStep1';
import CauseStep2 from './causeStep2';
import TableStep from './tableStep';
import ChartStep from './chartStep';

interface CauseProps {
    data: any;
}

function CauseStep({ data }: CauseProps) {
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

export default CauseStep;
