import React, { useEffect, useRef } from 'react';
import * as d3 from 'd3';
import _ from 'lodash';
import './index.less'

import nodeGreen from './hexagon-green.png';
import nodeRed from './hexagon-red.png';
import { formatUnit } from '@/services/util';

interface IProps {
    data: any
}
function SingleLinkTopology({ data }: IProps) {
    const chartRef = useRef<any>();

    const getNodeOrder = ({ nodes, edges }) => {
        let allNodeIds = nodes.map(node => node.id);
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
        const maxTime = _.chain(data.nodes).map(node => [node.selfP90, node.selfTime, (node.p90 - node.selfP90), (node.totalTime - node.selfTime)]).flatten().max().value();

        const WIDTH = chartRef.current.clientWidth;

        const svg = d3.select('#nodeChart').append('svg').style('width', WIDTH).style('height', data.nodes.length * 150 + 20);
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
        arrowheads.append('path').attr('d', 'M 0 0 L 10 5 L 0 10 L 4 5 z').attr('fill', '#454545');
        

        function drawNode(node, idx) {
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
                .attr('class', 'node_name')
                .attr('x', 70)
                .attr('y', 50);
                
            const timeY = 100;
            const timeX = 100;

            const selfTimeMaxWidth = (WIDTH - timeX) * 0.4
            const otherTimeMaxWidth = (WIDTH - timeX) * 0.6
            const timeTextYOffest = 15;

            nodeG.append("line")
                .attr('class', 'time_line')
                .attr("x1", timeX + selfTimeMaxWidth + 1)
                .attr("y1", timeY - 20)  
                .attr("x2", timeX + selfTimeMaxWidth + 1)
                .attr("y2", timeY + 65);

            nodeG.append('text')
                .text('selfTime')
                .attr('style', 'font-size: 11px')
                .attr('x', function() {
                    let textlen = this.getComputedTextLength();
                    return timeX + selfTimeMaxWidth - textlen - 5;
                })
                .attr('y', timeY - 10);

            nodeG.append('text')
                .text('otherTime')
                .attr('style', 'font-size: 11px')
                .attr('x', timeX + selfTimeMaxWidth + 7)
                .attr('y', timeY - 10);
            
            // 绘制P90
            nodeG.append('text')
                .text('P90')
                .attr('class', 'time_title')
                .attr('x', 70)
                .attr('y', timeY + timeTextYOffest);

            const selfP90Width = selfTimeMaxWidth * (node.selfP90 / maxTime);
            const p90OtherWidth = otherTimeMaxWidth * ((node.p90 - node.selfP90) / maxTime);
            nodeG.append('rect')
                .attr('class', 'time_rect')
                .attr('width', selfP90Width < 2 ? 2 : selfP90Width)
                .attr('height', 20)
                .attr('fill', '#5fe061')
                .attr('x', timeX + selfTimeMaxWidth - selfP90Width - 2)
                .attr('y', timeY);
            nodeG.append('text')
                .text(formatUnit(node.selfP90, 'ns'))
                .attr('class', 'time_value')
                .attr('x', function() {
                    let textlen = this.getComputedTextLength();
                    return timeX + selfTimeMaxWidth - textlen - 4
                })
                .attr('y', timeY + timeTextYOffest - 2);
            if (node.p90 - node.selfP90 > 0) {
                nodeG.append('rect')
                    .attr('class', 'time_rect')
                    .attr('width', p90OtherWidth < 2 ? 2 : p90OtherWidth)
                    .attr('height', 20)
                    .attr('fill', '#5fe061')
                    .attr('x', timeX + selfTimeMaxWidth + 4)
                    .attr('y', timeY);
                nodeG.append('text')
                    .text(formatUnit(node.p90 - node.selfP90, 'ns'))
                    .attr('class', 'time_value')
                    .attr('x', timeX + selfTimeMaxWidth + 6)
                    .attr('y', timeY + timeTextYOffest - 2);
            }
            // 绘制故障时间
            const selfWidth = selfTimeMaxWidth * (node.selfTime / maxTime);
            const OtherWidth = otherTimeMaxWidth * ((node.totalTime - node.selfTime) / maxTime);
            nodeG.append('text')
                .text('故障')
                .attr('class', 'time_title')
                .attr('x', 70)
                .attr('y', timeY + 34 + timeTextYOffest);
            nodeG.append('rect')
                .attr('class', 'time_rect')
                .attr('width', selfWidth < 2 ? 2 : selfWidth)
                .attr('height', 20)
                .attr('fill', '#ff3c3c')
                .attr('x', timeX + selfTimeMaxWidth - selfWidth - 2)
                .attr('y', timeY + 32);
            nodeG.append('text')
                .text(formatUnit(node.selfTime, 'ns'))
                .attr('class', 'time_value')
                .attr('x', function() {
                    let textlen = this.getComputedTextLength();
                    return timeX + selfTimeMaxWidth - textlen - 4
                })
                .attr('y', timeY + 32 + timeTextYOffest);
            if (node.totalTime - node.selfTime > 0) {
                nodeG.append('rect')
                    .attr('class', 'time_rect')
                    .attr('width', OtherWidth < 2 ? 2 : OtherWidth)
                    .attr('height', 20)
                    .attr('fill', '#ff3c3c')
                    .attr('x', timeX + selfTimeMaxWidth + 4)
                    .attr('y', timeY + 32);
                nodeG.append('text')
                    .text(formatUnit(node.totalTime - node.selfTime, 'ns'))
                    .attr('class', 'time_value')
                    .attr('x', timeX + selfTimeMaxWidth + 6)
                    .attr('y', timeY + 32 + timeTextYOffest);
            }            
        }

        function drawEdge(edge, idx) {
            let points: any[] = [];
            // @ts-ignore
            let nodeWarpRect = d3.select('#nodeChart').node()!.getBoundingClientRect();
            // @ts-ignore
            let sNodeRect: any = d3.select(`#node${nodeOrder.indexOf(edge.source)}`).node()!.getBoundingClientRect();
            // @ts-ignore
            let tNodeRect: any = d3.select(`#node${nodeOrder.indexOf(edge.target)}`).node()!.getBoundingClientRect();
            // console.log(nodeWarpRect, sNodeRect, tNodeRect);

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
                .attr('style', 'opacity: 1;stroke-width: 1;stroke: #333')
                .attr('fill', 'none')
                .attr('d', dpath)
                .attr('marker-end', 'url(#arrow)');
        }

        // 绘制node节点
        _.forEach(nodeOrder, (id, idx) => {
            let node = _.find(data.nodes, {id: id});
            drawNode(node, idx);
        })
        _.forEach(data.edges, (edge, idx) => {
            drawEdge(edge, idx);
        })
    }

    useEffect(() => {
        if (Object.keys(data).length > 0) {
            draw();
        }
    }, [data]);

    return (
        <div ref={chartRef} id='nodeChart' style={{ width: '100%', height: '100%', overflowY: 'auto', overflowX: 'hidden' }}></div>
    );
}
export default SingleLinkTopology;