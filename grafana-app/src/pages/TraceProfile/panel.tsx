import React, { useEffect } from 'react';
import { PanelProps } from '@grafana/data';

import './index.scss';

export interface PanelOptions {
    layout: string;
}
  
interface Props extends PanelProps<PanelOptions> {}
function TracePanel(props: Props) {
    const { options, data } = props;
    console.log(options, data);

    
    useEffect(() => {
        console.log('TopologyPanel Load');
    }, []);

    return (
        <div className='topology_warp'>
            <div>topology demo</div>
        </div>
    );
}

export default TracePanel;

