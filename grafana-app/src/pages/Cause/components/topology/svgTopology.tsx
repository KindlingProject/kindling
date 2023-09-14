import React, { useEffect, useRef } from 'react';
import { useTheme2 } from '@grafana/ui';
import * as d3 from 'd3';
import _ from 'lodash';

import nodeGreen from '../../../../img/hexagon-green.png';
import nodeRed from '../../../../img/hexagon-red.png';
import { formatUnit } from '../../../../utils/utils.format';

interface Props {
    data: any
}
function SingleLinkTopology({ data }: Props) {
    const theme = useTheme2();
    console.log(theme);
    const chartRef = useRef<any>();


    const getNodeOrder = ({ nodes, edges }: any) => {
        let allNodeIds = nodes.map((node: any) => node.id);
        // 构造根据edge的调用顺序从头到底排序的nodeId数组
        let nodeIdByEdge: string[] = [];
        allNodeIds.forEach((id: string) => {
            if (_.findIndex(edges, {target: id}) === -1) {
                nodeIdByEdge.push(id);
            }
        });
        while(_.findIndex(edges, {source: nodeIdByEdge[nodeIdByEdge.length - 1]}) > -1) {
            let edge = _.find(edges, {source: nodeIdByEdge[nodeIdByEdge.length - 1]});
            nodeIdByEdge.push(edge.target);
        }
        return nodeIdByEdge;
    }
    const draw = () => {
        const container: any = document.getElementById('nodeChart');
        if (data.nodes.length === 0) {
            container.innerHTML = ''
            return
        }
        container.innerHTML = ''

        const nodeOrder = getNodeOrder(data);
        const maxTime = _.chain(data.nodes).map(node => [node.p90, node.totalTime]).flatten().max().value();
        const WIDTH = chartRef.current.clientWidth;

        const svg = d3.select('#nodeChart').append('svg').style('width', WIDTH).style('height', data.nodes.length * 150);
        const g = svg.append('g').attr('id', 'chart_warp');

        const nodeWarp = g.append('g').attr('id', 'node_warp');
        const LineWarp = g.append('g').attr('id', 'line_warp');

        const defs = LineWarp.append('defs'); // defs定义可重复使用的元素
        const arrowheads = defs
            .append('marker') // 创建箭头
            .attr('id', 'arrow')
            .attr('markerUnits', 'strokeWidth')
            .attr('markerWidth', 16)
            .attr('markerHeight', 12)
            .attr('viewBox', '0 0 10 10')
            .attr('refX', 9)
            .attr('refY', 5)
            .attr('orient', 'auto');
        arrowheads.append('path').attr('d', 'M 0 0 L 10 5 L 0 10 L 4 5 z').attr('fill', theme.colors.text.primary);
        

        function drawNode(node: any, idx: number) {
            let nodeG = nodeWarp.append('g')
                .attr('class', 'cluster')
                .attr('transform', `translate(0, ${idx * 150})`);

            nodeG.append('image')
                .attr('id', `node${idx}`)
                .attr('class', 'node_image')
                .attr('x', 20)
                .attr('y', 30)
                .attr('width', 50)
                .attr('height', 50)
                .attr('xlink:href', () => {
                    return node.is_warn ? nodeRed : nodeGreen;
                });

            nodeG.append('text')
                .text(node.id)
                .style('fontSize', 12)
                .style('fill', theme.colors.text.primary)
                .attr('x', 70)
                .attr('y', 50);
                
            const timeY = 80;
            const timeX = 100;
            const timeMaxWidth = WIDTH - 100;
            const timeTextYOffest = 15;
            nodeG.append('text')
                .text('P90')
                .attr('style', `font-size: 12px; fill: ${theme.colors.text.primary}`)
                .attr('x', 70)
                .attr('y', timeY + timeTextYOffest);
            const p90Width = timeMaxWidth * (node.p90 / maxTime);
            const errWidth = timeMaxWidth * (node.totalTime / maxTime);
            nodeG.append('rect')
                .attr('class', 'time_rect')
                .attr('width', p90Width)
                .attr('height', 20)
                .attr('fill', '#5fe061')
                .attr('x', timeX)
                .attr('y', timeY);

            nodeG.append('text')
                .text(formatUnit(node.p90, 'ns'))
                .attr('style', `font-size: 12px; fill: ${theme.colors.text.primary}`)
                .attr('x', function () {
                    let textlen: any = this.getComputedTextLength();
                    if (p90Width < textlen + 20) {
                        return timeX + p90Width + 5;
                    } else {
                        return timeX + p90Width - textlen - 5;
                    }
                })
                .attr('y', timeY + timeTextYOffest);

            nodeG.append('text')
                .text('故障')
                .attr('style', `font-size: 12px; fill: ${theme.colors.text.primary}`)
                .attr('x', 70)
                .attr('y', timeY + 32 + timeTextYOffest);
            nodeG.append('rect')
                .attr('class', 'time_rect')
                .attr('width', errWidth)
                .attr('height', 20)
                .attr('fill', '#ff3c3c')
                .attr('x', timeX)
                .attr('y', timeY + 32);
            nodeG.append('text')
                .text(formatUnit(node.totalTime, 'ns'))
                .attr('style', `font-size: 12px; fill: ${theme.colors.text.primary}`)
                .attr('x', function () {
                    let textlen: any = this.getComputedTextLength();
                    if (errWidth < textlen + 20) {
                        return timeX + errWidth + 5;
                    } else {
                        return timeX + errWidth - textlen - 5;
                    }
                })
                .attr('y', timeY + 32 + timeTextYOffest);
            
        }

        function drawEdge(edge: any, idx: number) {
            let points: any[] = [];
            // @ts-ignore
            let nodeWarpRect = d3.select('#nodeChart').node()!.getBoundingClientRect();
            // @ts-ignore
            let sNodeRect: any = d3.select(`#node${nodeOrder.indexOf(edge.source)}`).node()!.getBoundingClientRect();
            // @ts-ignore
            let tNodeRect: any = d3.select(`#node${nodeOrder.indexOf(edge.target)}`).node()!.getBoundingClientRect();
            console.log(nodeWarpRect, sNodeRect, tNodeRect);

            let sx = sNodeRect.x - nodeWarpRect.x,
                tx = tNodeRect.x - nodeWarpRect.x,
                sy = sNodeRect.bottom - nodeWarpRect.y,
                ty = tNodeRect.top - nodeWarpRect.y;
        
            points.push({x: sx + 25, y: sy}, { x: tx + 25, y: ty })

            const dpath = d3
                .line()
                .x((d: any) => d.x)
                .y((d: any) => d.y)
                .curve(d3.curveBasis)(points);
            d3.select('#line_warp')
                .append('path')
                .attr('id', `line${idx}`)
                .attr('class', 'line_path')
                .attr('style', `opacity: 1;stroke-width: 1;stroke: ${theme.colors.text.primary}`)
                .attr('fill', 'none')
                .attr('d', dpath)
                .attr('marker-end', 'url(#arrow)');
        }

        // 绘制node节点
        _.forEach(nodeOrder, (id, idx) => {
            let node = _.find(data.nodes, {id: id});
            drawNode(node, idx);
        })
        _.forEach(data.edges, (edge, idx: number) => {
            drawEdge(edge, idx);
        })
    }

    useEffect(() => {
        if (Object.keys(data).length > 0) {
            draw();
        }
    // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [data]);

    return (
        <div ref={chartRef} id='nodeChart' style={{ width: '100%', height: '100%' }}></div>
    );
}
export default SingleLinkTopology;
