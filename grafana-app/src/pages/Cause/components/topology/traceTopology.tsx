import React, { useEffect, useRef } from 'react';
import * as d3 from 'd3';
import _ from 'lodash';
import { formatUnit } from '../../../../utils/utils.format';
import { useStyles2 } from '@grafana/ui';
import { GrafanaTheme2 } from '@grafana/data';
import { css } from '@emotion/css';

const redRGB = {
    r: 255, 
    g: 60,
    b: 60
};
const greenRGB = {
    r: 95, 
    g: 224,
    b: 97
};
// 构造对应深浅的色系组
function rgba(rgb: any, alpha: number) {
    return `rgba(${rgb.r}, ${rgb.g}, ${rgb.b}, ${alpha})`; 
} 
function rgbaToHex(rgba: any) {
    let [r, g, b, a] = rgba.match(/\d+(\.\d+)?/g); 
    let format = (n: any) => {
      n = parseInt(n, 10);
      return n.toString(16).padStart(2, '0');
    }
    return `#${format(r)}${format(g)}${format(b)}${Math.round(a * 255).toString(16)}`;
}
// 平铺树状节点数据，构造唯一ID。节点id存在重复调用的情况
function getAllNodes(list: any, data: any, parentId: any) {
    _.forEach(data, (node, idx) => {
        node.key = `${parentId}_${idx}`;
        let newNode = {...node};
        delete newNode.children;
        list.push(newNode);

        if (node.children.length > 0) {
            getAllNodes(list, node.children, `${parentId}_${idx}`);
        }
    });
}

function TraceTopology({data}: any) {
    // TODO 需要根据grafana的theme主题动态改变颜色(判断是否为暗黑主题)
    const style = useStyles2(getStyles);
    const nodes = useRef<any>();
    const maxandminMutatedValue = useRef<any>();
    const chartRef = useRef<any>();
    const svgRef = useRef<any>();
    const timeScale = useRef<any>();
    const colorScale = useRef<any>();

    const draw = (nodeList: any[], chartWarp: any, level: number) => {
        let nodeLevelG: any;
        if (d3.select(`#node_level_${level}`).node()) {
            nodeLevelG = d3.select(`#node_level_${level}`);
        } else {
            nodeLevelG = chartWarp.append('g').attr('id', `node_level_${level}`).attr('transform', `translate(0, ${level * 20})`);
        }
        
        if (_.some(nodeList, node => node.children && node.children.length > 0)) {
            level++;
        }

        nodeList.forEach(node => {
            const left = timeScale.current(node.startTime) as number;
            const right = timeScale.current(node.startTime + node.totalTime) as number;
            const rectWidth = right - left;
            const rect = nodeLevelG.append('rect')
                .style('cursor', 'pointer')
                .attr('width', rectWidth)
                .attr('data-key', node.key)
                .attr('height', 20)
                .attr('fill', () => {
                    if (node.mutatedValue > 0) {
                        let precent = node.mutatedValue / maxandminMutatedValue.current[1];
                        return colorScale.current(precent * 100)
                    } else {
                        let precent = -node.mutatedValue / maxandminMutatedValue.current[0];
                        return colorScale.current(precent * 100)
                    }
                })
                .attr('x', left)
                .attr('y', 5);
            
            nodeLevelG.append('text')
                .text(node.id.substring(0, Math.floor(rectWidth / 10)))
                .attr('class', 'node-text')
                .attr('style', 'font-size: 12px; fill: #efefef; pointer-events: none;')
                .attr('x', left + 5)
                .attr('y', 20);

            rect.on('click', function(this: any, e: any) {
                e.stopPropagation();
                const { pageX, pageY } = e;
                const key = d3.select(this).attr('data-key');
                let node = nodes.current.find((opt: any) => opt.key === key);
                const bodyWidth = document.body.clientWidth;
                if (pageX < bodyWidth / 2) {
                    d3.select('#chart_tooltip').attr('style', `visibility: visible;top: ${pageY + 10}px; left: ${pageX}px`);
                } else {
                    d3.select('#chart_tooltip').attr('style', `visibility: visible;top: ${pageY + 10}px; right: ${bodyWidth - pageX}px`);
                }

                const htmlContent = `<div class='tooltip_content'>
                    <div class='title'>${node.id}</div>
                    <div class='content'>
                        <div class='content-info'>URL: ${node.url}</div>
                        <div class='content-info-row'>
                            <div class='content-info'>是否被Profiled: ${node.isProfiled ? '是' : '否'}</div>
                            <div class='content-info'>自身耗时: ${formatUnit(node.selfTime, 'ns')}</div>
                        </div>
                        <div class='content-info-row'>
                            <div class='content-info'>历史P90: ${formatUnit(node.p90, 'ns')}</div>
                            <div class='content-info'>本次调用: ${formatUnit(node.totalTime, 'ns')}</div>
                        </div>
                        <div class='content-info-row'>
                            <div class='content-info'>突变值占比: ${(node.mutatedValuePrecent * 100).toFixed(2) + '%'}</div>
                            <div class='content-info'>突变值占比排名: ${node.mutatedValuePrecentSort}</div>
                        </div>
                    </div>
                </div>`
                d3.select('#chart_tooltip').html(htmlContent);
            })
            // rect.on('mouseenter', () => {
            //     console.log('mouseenter')
            // })
            // rect.on('mouseleave', () => {
            //     console.log('mouseleave')
            // })

            if (node.children && node.children.length > 0) {
                draw(node.children, chartWarp, level);
            } else {
                const height = chartRef.current.clientHeight;
                let contentHeight = (level + 1) * 20 + 70;
                if (contentHeight > height) {
                    svgRef.current.style('height', contentHeight);
                }
            }
        });
    }

    useEffect(() => {
        const nodeList: any[] = [];
        getAllNodes(nodeList, [data], 0);
        let maxMutatedValue = _.chain(nodeList).map('mutatedValue').filter(v => v > 0).sum().value();
        let minMutatedValue = _.chain(nodeList).map('mutatedValue').filter(v => v < 0).sum().value();
        maxandminMutatedValue.current = [minMutatedValue, maxMutatedValue];
        _.forEach(nodeList, (node: any) => {
            if (node.mutatedValue > 0) {
                node.mutatedValuePrecent = node.mutatedValue / maxandminMutatedValue.current[1];
            } else if (node.mutatedValue === 0) {
                node.mutatedValuePrecent = 0;
            } else {
                node.mutatedValuePrecent = -node.mutatedValue / maxandminMutatedValue.current[0];
            }
        });
        let mutatedValuePrecentList: number[] = _.chain(nodeList).sortBy('mutatedValuePrecent').map('mutatedValuePrecent').reverse().sortedUniq().value();
        _.forEach(nodeList, (node: any) => {
            node.mutatedValuePrecentSort = mutatedValuePrecentList.indexOf(node.mutatedValuePrecent) + 1;
        });
        nodes.current = nodeList;

        const container: any = document.getElementById('trace_topo');
        container.innerHTML = '';
        if (_.isEmpty(data)) {
            return
        }

        const WIDTH = chartRef.current.clientWidth;
        const HEIGHT = chartRef.current.clientHeight;
        timeScale.current = d3.scaleLinear()
            .domain([data.startTime, data.startTime + data.totalTime])
            .range([0, WIDTH]);

        const svg = d3.select('#trace_topo').append('svg').style('width', WIDTH).style('height', HEIGHT - 20);
        const g = svg.append('g').attr('id', 'chart_warp').attr('transform', `translate(0, 60)`);
        svgRef.current = svg;

        // 生成颜色对应比例尺
        const values = d3.range(-100, 120, 20);
        console.log(values);
        let colors: string[] = [];
        let alphavalue = -1;
        values.forEach((value, idx) => {
            let alpha = alphavalue + idx * 0.2;
            if (value < 0) {
                colors.push(rgbaToHex(rgba(greenRGB, Math.abs(alpha))));
            } else if (value === 0) {
                colors.push('#adabab');
            } else {
                colors.push(rgbaToHex(rgba(redRGB, alpha)));
            }
        });
        colorScale.current = d3.scaleLinear()
            .domain(values) 
            // @ts-ignore
            .range(colors)
        
        const colorRectWidth = 70;
        const colorGWarp = svg.append('g').attr('transform', `translate(${WIDTH/2 - values.length*colorRectWidth/2}, 10)`);
        colorGWarp.append('text').text('(-)耗时突变减少').attr('x', values.length*colorRectWidth/2 - 280).attr('y', 10).attr('style', 'font-size: 14px;fill: #5fe061')
        colorGWarp.append('text').text('(+)耗时突变增加').attr('x', values.length*colorRectWidth/2 + 170).attr('y', 10).attr('style', 'font-size: 14px;fill: #ff3c3c')
        colorGWarp.selectAll("rect")
            .data(values)
            .enter()
            .append("rect")
            .attr("x", (d, i) => i * colorRectWidth) 
            .attr("y", 20)
            .attr("width", colorRectWidth) 
            .attr("height", 20)
            .style("fill", d => colorScale.current(d));

        colorGWarp.selectAll(".color_text")
            .data(values)
            .enter()
            .append("text")
            .text((d) =>  d + '%')
            .attr('class', 'color_text')
            .attr("x", function(d, i) {
                return (i * colorRectWidth + colorRectWidth / 2) - this.getComputedTextLength() / 2
            }) 
            .attr("y", 35)
            .style("fill", '#efefef')

        svg.on('click', () => {
            d3.select('#chart_tooltip').attr('style', `visibility: none;`);
        })

        draw([data], g, 0);
    // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [data]);

    return (
        <div className={style.trace_topo_warp}> 
            <div ref={chartRef} id="trace_topo" style={{ width: '100%', height: '100%' }}></div>
            <div id="chart_tooltip" className={style.custom_tooltip}></div>
        </div>
    )
}

const getStyles = (theme: GrafanaTheme2) => ({
    trace_topo_warp: css`
        position: relative;
        width: 100%;
        height: 400px;
        overflow-y: auto;
        overflow-x: hidden;
    `,
    custom_tooltip: css`
        visibility: hidden;
        position: fixed;
        min-width: 200px;
        max-width: 400px;
        min-height: 80px;
        max-height: 200px;
        padding: 10px;
        background-color: #FFFFFF;
        border: 1px solid #dcdcdc;
        border-radius: 5px;
        font-size: 12px;
        color: #333333;
        .tooltip_content {
            width: 100%;
            height: 100%;
        }
        .title {
            padding-bottom: 5px;
            border-bottom: 1px solid #dcdcdc;
            overflow: hidden;
            text-overflow: ellipsis;
        }
        .content-info-row {
            display: flex;
            align-items: center;
            .content-info {
                width: 50%;
                min-width: 125px;
            }
        }
        .content-info {
            &:first-child {
                margin-top: 5px;
            }
        }
    `,
});

export default TraceTopology;
