import React, { ReactElement, ReactHTMLElement, ReactNode, useEffect, useRef } from 'react';
import G6 from '@antv/g6';
import './customNode';

interface IProps {
    data: any
}
function CallChart({ data }: IProps) {
    const sgraph = useRef<any>();
    
    const draw = () => {
        const container: any = document.getElementById('call_topo');
        console.log(data);
        if (data.nodes.length === 0) {
            container.innerHTML = ''
            return
        }
        container.innerHTML = ''

        const width = container.scrollWidth;
        const height = container.scrollHeight || 500;
        const graph = new G6.Graph({
            // renderer: 'svg',
            container: container,
            width,
            height,
            fitCenter: true,
            fitView: true,
            modes: {
                default: [
                    {
                        type: 'zoom-canvas',
                        minZoom: 0.8,
                        maxZoom: 2
                    }, 
                    'drag-canvas', 
                    'drag-node'
                ],
            },
            layout: {
                type: 'dagre',
                rankdir: 'TB',
                nodesepFunc: () => 80,
                ranksepFunc: () => 50,
            },
            defaultNode: {
                type: 'call-node',
            },
            defaultEdge: {
                size: 1,
                color: '#e2e2e2',
                style: {
                    endArrow: true,
                    lineWidth: 2,
                    stroke: '#C2C8D5',
                },
            }
        });
        graph.data(data);
        graph.render();

        sgraph.current = graph;
        graph.on('node-time-text:click', (evt: any) => {
            const node = evt.item.getModel();
            console.log(node);
        });
    }

    useEffect(() => {
        if (Object.keys(data).length > 0) {
            draw();
        }
    }, [data]);

    const changeSize = () => {
        const container = document.getElementById('call_topo');
        const width = container?.clientWidth;
        const height = container?.clientHeight;
        sgraph.current.changeSize(width, height);
        sgraph.current.fitCenter();
    }

    useEffect(() => {
        let MutationObserver = window.MutationObserver;
        let element: any = document.getElementById('left_chart_warp');
        let observer: any = new MutationObserver((mutationList) => {
            changeSize();
        });
        observer.observe(element, { attributes: true, attributeFilter: ['style', 'class'], attributeOldValue: true });
        return () => {
            observer = null;
        }
    }, [])

    return (
        <div id='call_topo' style={{ width: '100%', height: '100%' }}></div>
    );
}

export default CallChart;
