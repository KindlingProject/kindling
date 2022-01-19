import React, { useRef, useState, useEffect } from 'react';
import * as G6 from '@antv/g6';
import _ from 'lodash';
import './topology/node';
// import './topology/edge';
import { formatTime, formatCount, formatKMBT, formatPercent, nodeTooltip } from './topology/tooltip';
import { nsRelationHandle, detailRelationHandle, detailNodesHandle, detailEdgesHandle } from './topology/services'; 
import { PanelProps } from '@grafana/data';
import { SimpleOptions } from 'types';
import { css, cx } from 'emotion';
import { stylesFactory, Select } from '@grafana/ui';
// ErrorBoundary, Alert, Checkbox

interface VolumeProps {
  maxSentVolume: number; 
  maxReceiveVolume: number;
}
const meetricList: Array<{label: string; value: any; description?: string}> = [
  { label: 'Latency', value: 'latency' },
  { label: 'Calls', value: 'calls' },
  { label: 'Error Rate', value: 'errorRate' },
  { label: 'Sent Volume', value: 'sentVolume' },
  { label: 'Receive Volume', value: 'receiveVolume' },
  { label: 'RTT', value: 'rtt' },
  { label: 'Retransmit', value: 'retransmit' }
];

let SGraph: any;
interface Props extends PanelProps<SimpleOptions> {}
export const TopologyPanel: React.FC<Props> = ({ options, data, width, height, replaceVariables }) => {
  let graphRef: any = useRef();
  // const theme = useTheme();
  const namespace = replaceVariables('$namespace');
  const workload = replaceVariables('$workload');
  const styles = getStyles();
  // const [showCheckout] = useState<boolean>(namespace.split(',').length === 1);
  // const [showService, setShowService] = useState<boolean>(false);
  const [lineMetric, setLineMetric] = useState<any>('latency');
  const [volumes, setVolumes] = useState<VolumeProps>({maxSentVolume: 0, maxReceiveVolume: 0});

  console.log(namespace, workload);
  console.log(data);

  const crossLine = (edges: any, edgeModel: any, idx: number) => {
    edges.forEach((edge: any, edgeIdx: number) => {
      let source = edge.getSource().getModel();
      let target = edge.getTarget().getModel();
      if (source.id === edgeModel.target && target.id === edgeModel.source) {
        if (edgeIdx < idx) {
          edgeModel.labelCfg.refY = -10;
        }
      }
    });
  }
  // 根据当前指标选择更新边的样式
  const updateLinesAndNodes = (metric = lineMetric) => {
    const nodes = SGraph.getNodes();
    const edges = SGraph.getEdges();
    if (metric === 'latency' || metric === 'rtt' || metric === 'errorRate') {
      edges.forEach((edge: any, idx: number) => {
        let edgeModel = edge.getModel();
        let color: string;
        if (metric === 'latency') {
          color = edgeModel.latency > 1000 ? '#ff4c4c' : (edgeModel.latency > 200 ? '#f3ff69' : '#C2C8D5');
          edgeModel.label = formatTime(edgeModel.latency);
        } else if (metric === 'rtt') {
          color = edgeModel.rtt > 200 ? '#ff4c4c' : (edgeModel.rtt > 100 ? '#f3ff69' : '#C2C8D5');
          edgeModel.label = formatTime(edgeModel.latency);
        } else {
          color = edgeModel.errorRate > 0 ? '#ff4c4c' : '#C2C8D5';
          edgeModel.label = formatPercent(edgeModel.errorRate);
        }
        crossLine(edges, edgeModel, idx);
        edgeModel.style.stroke = color;
        edgeModel.style.lineWidth = 1;
        SGraph.updateItem(edge, edgeModel);
      });
      nodes.forEach((node: any) => {
        let nodeModel = node.getModel();
        if (metric === 'latency') {
          nodeModel.status = nodeModel.latency > 1000 ? 'red' : (nodeModel.latency > 200 ? 'yellow' : 'green');
        } else if (metric === 'rtt') {
          nodeModel.status = 'green';
        } else {
          nodeModel.status = nodeModel.errorRate > 0 ? 'red' : 'green';
        }
        SGraph.updateItem(node, nodeModel);
      });
    } else if (metric === 'sentVolume' || metric === 'receiveVolume'){
      if (metric === 'sentVolume') {
        let volumeStep = volumes.maxSentVolume / 5;
        if (edges.length === 1) {
          let edge = edges[0];
          let edgeModel = edge.getModel();
          edgeModel.style.lineWidth = 1;
          edgeModel.style.stroke = '#C2C8D5';
          edgeModel.label = formatKMBT(edgeModel.sentVolume);
          SGraph.updateItem(edge, edgeModel);
        } else {
          edges.forEach((edge: any, idx: number) => {
            let edgeModel = edge.getModel();
            let step = Math.floor(edgeModel.sentVolume / volumeStep);
            edgeModel.style.lineWidth = step === 0 ? 1 : 1.5 * step;
            edgeModel.style.stroke = '#C2C8D5';
            edgeModel.label = formatKMBT(edgeModel.sentVolume);
            crossLine(edges, edgeModel, idx);
            SGraph.updateItem(edge, edgeModel);
          });
        }
      } else {
        let volumeStep = volumes.maxReceiveVolume / 5;
        if (edges.length === 1) {
          let edge = edges[0];
          let edgeModel = edge.getModel();
          edgeModel.style.lineWidth = 1;
          edgeModel.style.stroke = '#C2C8D5';
          edgeModel.label = formatKMBT(edgeModel.receiveVolume);
          SGraph.updateItem(edge, edgeModel);
        } else {
          edges.forEach((edge: any, idx: number) => {
            let edgeModel = edge.getModel();
            let step = Math.floor(edgeModel.receiveVolume / volumeStep);
            edgeModel.style.lineWidth = step === 0 ? 1 : 1.5 * step;
            edgeModel.style.stroke = '#C2C8D5';
            edgeModel.label = formatKMBT(edgeModel.receiveVolume);
            crossLine(edges, edgeModel, idx);
            SGraph.updateItem(edge, edgeModel);
          });
        }
      }
      nodes.forEach((node: any) => {
        let nodeModel = node.getModel();
        nodeModel.status = 'green';
        SGraph.updateItem(node, nodeModel);
      });
    } else {
      edges.forEach((edge: any) => {
        let edgeModel = edge.getModel();
        edgeModel.style.stroke = '#C2C8D5';
        edgeModel.style.lineWidth = 1;
        edgeModel.label = formatCount(edgeModel[metric]);
        SGraph.updateItem(edge, edgeModel);
      });
      nodes.forEach((node: any) => {
        let nodeModel = node.getModel();
        nodeModel.status = 'green';
        SGraph.updateItem(node, nodeModel);
      });
    }
  }
  // 绘制拓扑图
  const draw = (gdata: any) => {
    const inner: any = document.getElementById('kindling_topo');
    inner.innerHTML = '';

    const graph = new G6.Graph({
      // renderer: 'svg',
      container: 'kindling_topo',
      width: width,
      height: height,
      fitCenter: true,
      autoPaint: false,
      plugins: [nodeTooltip],
      modes: {
        default: [
          {
            type: 'drag-canvas',
            enableOptimize: true,
          }, {
            type: 'zoom-canvas',
            maxZoom: 1.8,
            minZoom: 0.4
          }, 
          'drag-node'
        ]
      },
      layout: {
        type: 'dagre',
        rankdir: 'LR',
        align: 'DL',
        // controlPoints: true,
        // workerEnabled: nodeNum > 200
      },
      defaultNode: {
        type: 'custom-node'
      },
      // defaultEdge: {
      //   type: 'service-edge'
      // },
      defaultEdge: {
        size: 1,
        type: 'cubic-horizontal',
        label: '',
        labelCfg: {
          refY: 10,
          autoRotate: true,
          style: {
            fontWeight: 400,
            fill: '#C2C8D5',
          }
        },
        style: {
          radius: 10,
          offset: 5,
          endArrow: true,
          lineWidth: 1,
          stroke: '#C2C8D5',
        }
      }
    });
    graph.data(gdata);
    graph.render();

    SGraph = graph;
    updateLinesAndNodes();
  };
  // 初始化数据处理：生成拓扑数据，获取调用关系流量最大值
  const initData = () => {
    // 处理grafana查询数据生成对应的拓扑调用数据结构
    let nodes: any[] = [], edges: any[] = [];
    let topoData: any = _.filter(data.series, (item: any) => item.refId === 'A');
    let edgeTimeData: any = _.filter(data.series, (item: any) => item.refId === 'I');
    let edgeSendVolumeData: any = _.filter(data.series, (item: any) => item.refId === 'B');
    let edgeReceiveVolumeData: any = _.filter(data.series, (item: any) => item.refId === 'C');
    let edgeRetransmitData: any = _.filter(data.series, (item: any) => item.refId === 'J');
    let edgeRTTData: any = _.filter(data.series, (item: any) => item.refId === 'K');
    const edgeData = {
      edgeCallData: topoData,
      edgeTimeData,
      edgeSendVolumeData,
      edgeReceiveVolumeData,
      edgeRetransmitData,
      edgeRTTData
    };
    
    let nodeCallsData: any = _.filter(data.series, (item: any) => item.refId === 'D'); // 次数调用增长
    let nodeTimeData: any = _.filter(data.series, (item: any) => item.refId === 'E');
    let nodeErrorRateData: any = _.filter(data.series, (item: any) => item.refId === 'F');
    let nodeSendVolumeData: any = _.filter(data.series, (item: any) => item.refId === 'G');
    let nodeReceiveVolumeData: any = _.filter(data.series, (item: any) => item.refId === 'H');
    const nodeData = {
      nodeCallsData,
      nodeTimeData,
      nodeErrorRateData,
      nodeSendVolumeData,
      nodeReceiveVolumeData
    };
    console.log('edgeData', edgeData);
    console.log('nodeData', nodeData);
    // 当namespace为all的时候，只绘制对应namespace的调用关系
    if (namespace.indexOf(',') > -1) {
      let result: any = nsRelationHandle(topoData, nodeData, edgeData);
      nodes = [].concat(result.nodes);
      edges = [].concat(result.edges);
    } else {
      // 当workload为all的时候，只绘制对应namespace下所有workload的调用关系
      if (workload.indexOf(',') > -1) {
        let result = _.filter(topoData, (item: any) => item.fields[1].labels.dst_namespace === namespace || item.fields[1].labels.src_namespace === namespace);
        console.log('workload Topology', result);
        _.forEach(result, item => {
          let tdata: any = item.fields[1].labels;
          // let source: string, target: string;
          let { node: targetNode, target } = detailRelationHandle(nodes, edges, namespace, tdata, 'dst', false);
          let { node: sourceNode, source } = detailRelationHandle(nodes, edges, namespace, tdata, 'src', false);
          sourceNode && nodes.push(sourceNode);
          targetNode && nodes.push(targetNode);
          if (_.findIndex(edges, {source: source, target: target}) === -1) {
            edges.push({
              source: source,
              target: target
            });
          }
        });
        nodes = detailNodesHandle(nodes, nodeData);
        edges = detailEdgesHandle(nodes, edges, edgeData);
      } else {
        // 具体namespace和workload下的pod调用关系
        let result = _.filter(topoData, (item: any) => (item.fields[1].labels.dst_namespace === namespace && item.fields[1].labels.dst_workload_name === workload) || (item.fields[1].labels.src_namespace === namespace && item.fields[1].labels.src_workload_name === workload));
        console.log('pod Topology', result);
        _.forEach(result, item => {
          let tdata: any = item.fields[1].labels;
          let { node: targetNode, target } = detailRelationHandle(nodes, edges, namespace, tdata, 'dst', true, workload);
          let { node: sourceNode, source } = detailRelationHandle(nodes, edges, namespace, tdata, 'src', true, workload);
          sourceNode && nodes.push(sourceNode);
          targetNode && nodes.push(targetNode);
          if (_.findIndex(edges, {source: source, target: target}) === -1) {
            edges.push({
              source: source,
              target: target
            });
          }
        });
        nodes = detailNodesHandle(nodes, nodeData);
        edges = detailEdgesHandle(nodes, edges, edgeData);
      }
    }
    
    let gdata = {
      nodes: nodes,
      edges: edges
    }
    console.log(gdata);
    draw(gdata);

    let volumeData: VolumeProps = {
      maxSentVolume: _.max(_.map(gdata.edges, 'sentVolume')),
      maxReceiveVolume: _.max(_.map(gdata.edges, 'receiveVolume'))
    }
    setVolumes(volumeData);
  }

  useEffect(() => {
    if (data.state === 'Done') {
      initData();
    }
	// eslint-disable-next-line react-hooks/exhaustive-deps
  }, [data]);

  // 切换线段指标时，相应的线段样式更新
  const lineMetricChange = (opt: any) => {
    setLineMetric(opt.value);
    updateLinesAndNodes(opt.value);
  }
  // const changeShowService = () => {
  //   setShowService(!showService ? true : false);
  // }
  return (
    <div
      className={cx(
        styles.wrapper,
        css`
          width: ${width}px;
          height: ${height}px;
        `
      )}
    >
      <div className={styles.topLineMetric}>
        <div className={styles.metricSelect}>
          <span style={{ width: '180px' }}>Call Line Metric</span>
          <Select value={lineMetric} options={meetricList} onChange={lineMetricChange}/>
        </div>
        {/* {
          showCheckout ? <Checkbox css="" value={showService} onChange={changeShowService} label='View Service Call'/> : null
        } */}
      </div>
      <div id="kindling_topo" style={{ height: '100%' }} ref={graphRef}></div>
    </div>
  );
};

const getStyles = stylesFactory(() => {
  return {
    wrapper: css`
      position: relative;
    `,
    topLineMetric: css`
      position: absolute;
      top: 0;
      left: 0;
      z-index: 10;
      display: flex;
      flex-direction: column;
    `,
    metricSelect: css`
      display: flex;
      align-items: center;
      margin-bottom: 10px;
    `,
    svg: css`
      position: absolute;
      top: 0;
      left: 0;
    `,
    textBox: css`
      position: absolute;
      bottom: 0;
      left: 0;
      padding: 10px;
    `,
  };
});
