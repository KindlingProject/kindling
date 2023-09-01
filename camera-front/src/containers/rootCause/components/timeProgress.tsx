import React from 'react';
import { formatUnit } from '@/services/util';

interface IProps {
    p90: number;
    data: number; 
    width?: number
}
function TimeProgress({ p90, data }: IProps) {
    
    const widthPercent = p90 / data * 80;

    return (
        <div className='duration_warp'>
            <div className='duration_item'>
                <span>历史数据P90</span>
                {
                    p90 > 0 ? <React.Fragment>
                        <div className='health-line' style={{ width: `${widthPercent}%` }}></div>
                        <span>{ formatUnit(p90, 'ns') }</span>
                    </React.Fragment> : null
                }
            </div>
            <div className='duration_item'>
                <span>本次故障调用</span>
                {
                    data > 0 ? <React.Fragment>
                        <div className='abnormal-line'></div>
                        <span>{ formatUnit(data, 'ns') }</span>
                    </React.Fragment> : null
                }
            </div>
        </div>
    );
}

export default TimeProgress;
