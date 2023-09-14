import React, { useState, useEffect } from 'react';
import { PanelProps } from '@grafana/data';
import { Card, Button, Modal } from '@grafana/ui';
import _ from 'lodash';
import SingleLinkTopology from './components/topology/svgTopology';
import TraceTopology from './components/topology/traceTopology';
import CauseStep from './components/causeStep';
import CauseDataSource from './causeDataSource';
import { formatUnit } from '../../utils/utils.format';
import './index.scss';

const causeDataQuery = new CauseDataSource();
export interface CustomVizOptions {
    traceId: string;   
}
interface Props extends PanelProps<CustomVizOptions> {}
function CustomVizPanel(props: Props) {
    const { options } = props;
    // console.log(options, data);
    const { traceId } = options;
    
    const [entryTrace, setEntryTrace] = useState<any>({});
    const [topologyData, setTopologyData] = useState<any>({});
    const [topologyTreeData, setTopologyTreeData] = useState({});
    const [reportData, setReportData] = useState<any>({}); 
    const [modalVisible, setModalVisible] = useState<boolean>(false)
    
    const handleTopologyData = (nodes: any[], edges: any[], node: any) => {
        if (node.isPath) {
            let newNode: any = { ...node };
            delete newNode.children;
            nodes.push(newNode);
        }
        
        if (node.children && node.children.length > 0) {
            if (_.some(node.children, node => node.isPath)) {
                let subNode = _.find(node.children, {isPath: true});
                edges.push({ source: node.id, target: subNode.id });
                handleTopologyData(nodes, edges, subNode);
            } 
        }
    }
    useEffect(() => {
        console.log('CustomVizPanel Load');
        causeDataQuery.excuteQuery(`/query/mutatedtraces/${traceId}`).then(res => {
            console.log('res', res);
            if (res.success) {
                const nodes: any[] = [], edges: any[] = [];
                handleTopologyData(nodes, edges, res.data);
                // console.log(nodes, edges);
                let spanId = '';
                _.forEach(nodes, (node: any) => {
                    if (node.isMutated) {
                        node.is_warn = true;
                        spanId = node.spanId;
                    }
                });
                setTopologyTreeData(res.data);
                setEntryTrace(nodes[0]);
                setTopologyData({ nodes, edges });

                causeDataQuery.excuteQuery(`/flows/run/${traceId}/${spanId}`).then(res => {
                    setReportData(res);
                });
            }
        });
    // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [traceId]);

    const toggleTopologyModal = () => {
        setModalVisible(!modalVisible);
    }

    return (
        <div className='cause_warp'>
            <Card>
                <Card.Meta>{[`故障描述: ${entryTrace.contentKey}请求异常`, `故障发生时间: ${formatUnit(entryTrace.startTime / 1000000, 'date')}`, `请求耗时: ${formatUnit(entryTrace.totalTime, 'ns')}`]}</Card.Meta>
            </Card>
            <div className='report_content'>
                <div className='topology_warp'>
                    <Card>
                        <Card.Description>
                            <div className='chart_warp'>
                                {
                                    (topologyData.nodes && topologyData.nodes.length > 0) ? <Button size='xs' variant="primary" fill="text" className='chart_btn' onClick={toggleTopologyModal}>全链路调用</Button> : null
                                }
                                <SingleLinkTopology data={topologyData}/>
                            </div>
                        </Card.Description>
                    </Card>
                </div>
                <div className='report_warp'>
                    <Card>
                        <Card.Description>
                            <span>故障根因：</span>
                            <span style={{ color: '#ff3c3c' }}>{ reportData.conclusion?.message || '' }</span>
                        </Card.Description>
                    </Card>
                    {
                        reportData?.reports ? reportData.reports.map((item: any, idx: number) => <CauseStep data={item} key={idx} />) : null
                    }
                </div>
            </div>
            <Modal isOpen={modalVisible} title='全链路调用火焰图' onDismiss={toggleTopologyModal} className='trace_topology_modal'>
                <TraceTopology data={topologyTreeData}/>
            </Modal>
        </div>
        
    );
}

export default CustomVizPanel;

