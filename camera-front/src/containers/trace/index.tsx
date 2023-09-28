import { useEffect, useState, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { Input, Button } from 'antd';
import G6 from '@antv/g6';
import _ from 'lodash';
import { getTraceTopology } from '@/request';
import './customNode';
import './index.less';


function Trace() {
    const navigate = useNavigate();
    const [traceId, setTraceId] = useState('');
    // const [timestamp, setTimestamp] = useState('');

    const init = () => {
        const params = {
            traceId: traceId
        };
        getTraceTopology(params).then(res => {
            console.log(res);
            draw(res.data.data);
        });
    }

    const draw = (data) => {
        const container: any = document.getElementById('trace_topo');
        if (data.nodes.length === 0) {
            container.innerHTML = ''
            return
        }
        container.innerHTML = ''

        const width = container.scrollWidth;
        const height = container.scrollHeight || 500;
        const graph = new G6.Graph({
            container: container,
            width,
            height,
            fitCenter: true,
            // fitView: true,
            modes: {
                default: ['drag-canvas', 'drag-node'],
            },
            layout: {
                type: 'dagre',
                rankdir: 'LR',
                align: 'DL',
                nodesepFunc: () => 1,
                ranksepFunc: () => 100,
            },
            defaultNode: {
                type: 'custom-node',
            },
            defaultEdge: {
                size: 1,
                color: '#e2e2e2',
                style: {
                endArrow: {
                    path: 'M 0,0 L 8,4 L 8,-4 Z',
                    fill: '#e2e2e2',
                },
                },
            }
        });
        graph.data(data);
        graph.render();

        graph.on('node-time-text:click', (evt: any) => {
            const node = evt.item.getModel();
            const timeIdx = evt.target.attr('timeIdx');
            navigate(`/thread?query=es&pid=${node.pid}&stime=${node.list[timeIdx].endTime - 1}&etime=${node.list[timeIdx].endTime + 1}&protocl=${node.protocol}`);
        });
    }
    return (
        <div className='trace_warp'>
            <header className='trace_header'>
                <div>
                    <Input style={{ width: 320, marginRight: 10 }} value={traceId} onChange={(e) => setTraceId(e.target.value)} />
                    {/* <Input type="number" style={{ width: 180, marginRight: 10 }} value={timestamp} onChange={(e) => setTimestamp(e.target.value)} /> */}
                    <Button onClick={init}>查询</Button>   
                </div>
            </header>
            <div className='trace_body'>
                <div id='trace_topo' className='trace_content'></div>
            </div>
        </div>
    );
}

export default Trace;
