import React, { useRef, useState, useEffect } from 'react';
import * as G6 from '@antv/g6';
import _ from 'lodash';
import './topology/node';
import './topology/edge';
import { nodeTooltip } from './topology/tooltip';
import TopoLegend from './topology/legend';
import { metricList, layoutOptions, directionOptions, viewRadioOptions, showServiceOptions, NodeDataProps, EdgeDataProps, transformWorkload,
  transformData, getNodeTypes, nsRelationHandle, connFailNSRelationHandle, workloadRelationHandle, connFailWorkloadRelationHandle, TopologyProps, topoMerge } from './topology/services'; 
import { buildLayout, serviceLineUpdate, updateLinesAndNodes } from './topology/topology'
import { PanelProps } from '@grafana/data';
import { SimpleOptions } from 'types';
import { css, cx } from 'emotion';
import { stylesFactory, Select, RadioButtonGroup, Icon, Tooltip, Spinner } from '@grafana/ui';
import { locationService } from '@grafana/runtime';

interface VolumeProps {
  maxSentVolume: number; 
  maxReceiveVolume: number;
  minSentVolume: number; 
  minReceiveVolume: number;
}

let SGraph: any;
let topoData: any, connFailTopoData: any, nodeData: NodeDataProps, edgeData: EdgeDataProps;
let connFailTopo: TopologyProps;
interface Props extends PanelProps<SimpleOptions> {}
export const TopologyPanel: React.FC<Props> = ({ options, data, width, height, replaceVariables }) => {
  let graphRef: any = useRef();
  // const theme = useTheme();
  const namespace = replaceVariables('$namespace');
  const workload = replaceVariables('$workload');
  const [_location, setLocation] = useState(locationService.getLocation())

  const styles = getStyles();
  const [graphData, setGraphData] = useState<any>({}); 
  const [layout, setLayout] = useState<string>('dagre'); 
  const [loading, setLoading] = useState<boolean>(true); 
  const [showCheckbox, setShowCheckbox] = useState<boolean>(namespace === 'all');
  const [showService, setShowService] = useState<boolean>(false);
  const [showView, setShowView] = useState<boolean>(false);
  const [firstChangeDir, setFirstChangeDir] = useState<boolean>(false);
  const [direction, setDirection] = useState<string>('LR');
  const [view, setView] = useState<string>('workload_view');
  const [lineMetric, setLineMetric] = useState<any>('latency');
  const [volumes, setVolumes] = useState<VolumeProps>({maxSentVolume: 0, maxReceiveVolume: 0, minSentVolume: 0, minReceiveVolume: 0});
  const [nodeTypesList, setNodeTypesList] = useState<any[]>([]);

  // console.log(options, namespace, workload, width, height);
  // console.log(data);

  // draw topology
  const draw = (gdata: any, metric = lineMetric, serviceLine = showService) => {
    const inner: any = document.getElementById('kindling_topo');
    inner.innerHTML = '';
    let data = _.cloneDeep(gdata);
    const graph = new G6.Graph({
      // renderer: 'svg',
      container: 'kindling_topo',
      width: width - 240,
      height: height,
      // fitView: true,
      fitViewPadding: 10,
      maxZoom: 1.5,
      minZoom: 0.25,
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
            maxZoom: 1.5,
            minZoom: 0.25
          }, 
          'drag-node'
        ]
      },
      layout: buildLayout(layout, direction),
      defaultNode: {
        type: 'custom-node'
      },
      defaultEdge: {
        type: 'service-edge',
        labelCfg: {
          refY: serviceLine ? 15 : 10,
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
    graph.data(data);
    graph.render();
    /**
     * Close local rendering to solve the lingering problem of some browser drag nodes
     * 关闭局部渲染，解决部分浏览器拖拽节点的残影问题
     */
    graph.get('canvas').set('localRefresh', false);

    SGraph = graph;
    serviceLineUpdate(SGraph, direction);
    updateLinesAndNodes(SGraph, options, volumes, metric, serviceLine);
  };
 
  /**
   * When you go back to the topology drawing data, update the type array of the corresponding node and the value of Max and min of the flow on the side
   * 重新回去拓扑绘制数据时，更新对应的节点的类型数组和边上流量max、min的数值
   */
  const handleResult = (gdata: any) => {
    let typeList = getNodeTypes(gdata.nodes);
    console.log('nodeTypesList', typeList);
    setNodeTypesList(typeList);
    let volumeData: VolumeProps = {
      maxSentVolume: _.max(_.map(gdata.edges, 'sentVolume')),
      maxReceiveVolume: _.max(_.map(gdata.edges, 'receiveVolume')),
      minSentVolume: _.min(_.map(gdata.edges, 'sentVolume')),
      minReceiveVolume: _.min(_.map(gdata.edges, 'receiveVolume'))
    }
    setVolumes(volumeData);
  }
  const buildtopoData = () => {
    let nodes: any[] = [], edges: any[] = [];
    let connFailResult: TopologyProps = {
      nodes: [],
      edges: []
    };
    if (namespace.indexOf(',') > -1) {
      let result: any = nsRelationHandle(topoData, nodeData, edgeData);
      connFailResult = connFailNSRelationHandle(connFailTopoData);
      nodes = [].concat(result.nodes);
      edges = [].concat(result.edges);
    } else {
      let showPod = workload.indexOf(',') === -1;
      setView(showPod ? 'pod_view' : 'workload_view');
      let result: any = workloadRelationHandle(transformWorkload(workload), namespace, topoData, nodeData, edgeData, showPod, showService);
      connFailResult = connFailWorkloadRelationHandle(transformWorkload(workload), namespace, connFailTopoData, showPod, showService);
      nodes = [].concat(result.nodes);
      edges = [].concat(result.edges);
    }
    
    let gdata = {
      nodes: nodes,
      edges: edges
    }
    connFailTopo = _.cloneDeep(connFailResult);
    setGraphData(gdata);
    if (lineMetric === 'connFail') {
      const data = topoMerge(gdata, connFailTopo);
      draw(data);
    } else {
      draw(gdata);
    }
    handleResult(gdata);
  }
  // Initial data processing: Generates topology data
  const initData = () => {
    topoData = transformData(_.filter(data.series, (item: any) => item.refId === 'A'));
    connFailTopoData = transformData(_.filter(data.series, (item: any) => item.refId === 'L'));
    let edgeTimeData: any = transformData(_.filter(data.series, (item: any) => item.refId === 'I'));
    let edgeSendVolumeData: any = transformData(_.filter(data.series, (item: any) => item.refId === 'B'));
    let edgeReceiveVolumeData: any = transformData(_.filter(data.series, (item: any) => item.refId === 'C'));
    let edgeRetransmitData: any = transformData(_.filter(data.series, (item: any) => item.refId === 'J'));
    let edgeRTTData: any = transformData(_.filter(data.series, (item: any) => item.refId === 'K'));
    let edgePackageLostData = transformData(_.filter(data.series, (item: any) => item.refId === 'F'));
    edgeData = {
      edgeCallData: topoData,
      edgeTimeData,
      edgeSendVolumeData,
      edgeReceiveVolumeData,
      edgeRetransmitData,
      edgeRTTData,
      edgePackageLostData
    };
    
    let nodeCallsData: any = transformData(_.filter(data.series, (item: any) => item.refId === 'D'));
    let nodeTimeData: any = transformData(_.filter(data.series, (item: any) => item.refId === 'E'));
    let nodeSendVolumeData: any = transformData(_.filter(data.series, (item: any) => item.refId === 'G'));
    let nodeReceiveVolumeData: any = transformData(_.filter(data.series, (item: any) => item.refId === 'H'));
    nodeData = {
      nodeCallsData,
      nodeTimeData,
      nodeSendVolumeData,
      nodeReceiveVolumeData
    };
    // console.log('edgeData', edgeData);
    // console.log('nodeData', nodeData);

    buildtopoData();
  }

  useEffect(() => {
    if (SGraph) {
      if (lineMetric === 'connFail') {
        const data = topoMerge(graphData, connFailTopo);
        draw(data);
      } else {
        draw(graphData);
      }  
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [layout]);
  useEffect(() => {
    if (SGraph) {
      updateLinesAndNodes(SGraph, options, volumes, lineMetric, showService);
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [volumes]);

  useEffect(() => {
    buildtopoData();
    if (namespace.indexOf(',') === -1) {
      setShowCheckbox(true);
      if (workload.indexOf(',') === -1) {
        setShowView(true);
      } else {
        setShowView(false);
      }
    } else {
      setShowCheckbox(false);
      setShowView(false);
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [namespace, workload]);


  useEffect(() => {
    const history = locationService.getHistory()
    const unlisten = history.listen((location: any) => {
      console.log('location', location);
      setLocation(location)
    })
    return unlisten
  }, []);

  useEffect(() => {
    console.log(data);
    if (data.state === 'Done') {
      setLoading(false);
      initData();
    }
	// eslint-disable-next-line react-hooks/exhaustive-deps
  }, [data]);

  // When the segment indicator is switched, the corresponding segment style is updated
  const lineMetricChange = (opt: any) => {
    setLineMetric(opt.value);
    if (opt.value === 'connFail') {
      const data = topoMerge(graphData, connFailTopo);
      draw(data, opt.value);
    } else {
      if (lineMetric === 'connFail') {
        draw(graphData, opt.value);
      } else {
        updateLinesAndNodes(SGraph, options, volumes, opt.value, showService);
      }
    }    
  }
  const changeLayout = (opt: any) => {
    if (opt.value === layout) {
      return;
    }
    setLayout(opt.value);
    // draw(graphData);
  } 
  const changeDirection = (value: any) => {
    setDirection(value);
    SGraph.updateLayout({
      rankdir: value
    });
    SGraph.fitView(10);
    if (!firstChangeDir) {
      serviceLineUpdate(SGraph, value);
      setFirstChangeDir(true);
    }
  }
  // Whether to display service calls on invocation relationships
  const changeShowService = () => {
    let show = !showService ? true : false;
    setShowService(show);
    let { nodes, edges } = workloadRelationHandle(transformWorkload(workload), namespace, topoData, nodeData, edgeData, view === 'pod_view', show);
    let gdata = {
      nodes: nodes,
      edges: edges
    }
    draw(gdata, lineMetric, show);
    setGraphData(gdata);
    handleResult(gdata);
  }
  // toggle View Mode。Switch between the workload view and pod view
  const changeView = (value: any) => {
    setView(value);
    let { nodes, edges } = workloadRelationHandle(transformWorkload(workload), namespace, topoData, nodeData, edgeData, value === 'pod_view', showService);
    let gdata = {
      nodes: nodes,
      edges: edges
    }
    draw(gdata);
    setGraphData(gdata);
    handleResult(gdata);
  }

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
        {/* <div className={styles.filterWarp}>
          <div>
            <InlineField label="Namespace">
             <Select value={namespace} options={namespaceList} onChange={namespaceChange}/>
            </InlineField>
          </div>
          <div>
            <InlineField label="Workload">
             <Select value={workload} options={workloadList} onChange={workloadChange}/>
            </InlineField>
          </div>
        </div> */}
        <div className={styles.metricSelect}>
          <span style={{ width: '170px' }}>Call Line Metric</span>
          <Select value={lineMetric} options={metricList} onChange={lineMetricChange}/>
        </div>
      </div>
      <div className={styles.topRightWarp}>
        <div className={styles.viewRadioMode}>
          <div>
            <span>Layout</span>
            <Tooltip content="change topology layout。">
              <Icon name="question-circle" />
            </Tooltip>
          </div>
          <div style={{ width: 150 }}>
            <Select value={layout} options={layoutOptions} onChange={changeLayout}/>
          </div>
        </div>
        {
          layout === 'dagre' ? <div className={styles.viewRadioMode}>
            <div>
              <span>Layout Direction</span>
              <Tooltip content="change Dargre topology layout direction mean top to bottom，LR mean left to right。">
                <Icon name="question-circle" />
              </Tooltip>
            </div>
            <RadioButtonGroup options={directionOptions} value={direction} onChange={changeDirection}/>
          </div> : null
        }
        {
          showView ? <div className={styles.viewRadioMode}>
            <span>View Mode</span>
            <RadioButtonGroup options={viewRadioOptions} value={view} onChange={changeView}/>
          </div> : null
        }
        {
          showCheckbox ? <div className={styles.viewRadioMode}>
            <div>
              <span>Show Services</span>
              <Tooltip content="if the network communicate by Kubernetes service, the service name will be shown。">
                <Icon name="question-circle" />
              </Tooltip>
            </div>
            <RadioButtonGroup options={showServiceOptions} value={showService} onChange={changeShowService}/>
          </div> : null
        }
        <TopoLegend typeList={nodeTypesList} metric={lineMetric} volumes={volumes} options={options}/>
      </div>
      <div id="kindling_topo" style={{ height: '100%' }} ref={graphRef}></div>
      {
        loading ? <div className={styles.spinner_warp}>
          <Spinner className={styles.spinner_icon}/>
        </div> : null
      }
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
    filterWarp: css`
      display: flex;
      align-items: center;
      margin-bottom: 10px;
    `,
    metricSelect: css`
      display: flex;
      align-items: center;
      margin-bottom: 10px;
    `,
    topRightWarp: css`
      position: absolute;
      top: 0;
      right: 0;
      z-index: 10;
      display: flex;
      flex-direction: column;
      width: 245px;
    `,
    viewRadioMode: css`
      display: flex;
      align-items: center;
      justify-content: space-between;
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
    spinner_warp: css`
      position: absolute;
      top: 0;
      left: 0;
      width: 100%;
      height: 100%;
      background-color: #181b1fc2;
      z-index: 20;
    `,
    spinner_icon: css`
      position: absolute;
      font-size: xx-large;
      top: 48%;
      left: 49%;
    `,
  };
});
