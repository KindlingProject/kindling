import React, { useEffect, useState } from 'react';
import { Row, Col, Button, Space, Spin, Descriptions, Empty, message, Modal } from 'antd';
import _ from 'lodash';
import './index.less';
import TimeProgress from './components/timeProgress';
import CallChart from './components/callChart';
import TraceTopology from './components/svgTopology/traceTopology';
import SingleLinkTopology from './components/svgTopology';
import CauseStep from './components/causeStep';

import { formatUnit } from '@/services/util';
import { useSearchParams } from '@/services/hooks';
import { getSingleTraceTopology, getCauseReports } from '@/request';


function RootCause() {
    const { traceId } = useSearchParams();
    const [loading, setLoading] = useState<boolean>(false);
    const [topologyFullScreen, setTopologyFullScreen] = useState<boolean>(false);
    const [causeFullScreen, setCauseFullScreen] = useState<boolean>(false);
    const [entryTrace, setEntryTrace] = useState<any>({});
    const [topologyTreeData, setTopologyTreeData] = useState({});
    const [topologyData, setTopologyData] = useState<any>({});
    const [reportData, setReportData] = useState<any>({});
    const [topoModalVisible, setTopoModalVisible] = useState<boolean>(false);

    const handleTopologyData = (nodes, edges, node) => {
        if (node.isPath) {
            let newNode: any = { ...node };
            delete newNode.children;
            nodes.push(newNode);
        }
        
        if (node.children && node.children.length > 0) {
            console.log(_.some(node.children, node => node.isPath));
            if (_.some(node.children, node => node.isPath)) {
                let subNode = _.find(node.children, {isPath: true});
                edges.push({ source: node.id, target: subNode.id });
                handleTopologyData(nodes, edges, subNode);
            } 
        }
    }
    const getTopologyByTraceId = () => {
        setLoading(true);
        getSingleTraceTopology(traceId).then(res => {
            console.log(res);
            if (res.data.success) {
                const nodes = [], edges = [];
                setTopologyTreeData(res.data.data);
                handleTopologyData(nodes, edges, res.data.data);
                console.log(nodes, edges);
                let spanId = '';
                _.forEach(nodes, (node: any) => {
                    if (node.isMutated) {
                        node.is_warn = true;
                        spanId = node.spanId;
                    }
                });
                setEntryTrace(nodes[0]);
                setTopologyData({ nodes, edges });

                getReportsData(spanId);
            } else {
                message.warning(res.data.errorMsg);
            }
        }).finally(() => {
            setLoading(false);
        });
    }
    
    const getReportsData = (spanId) => {
        getCauseReports(traceId, spanId).then(res => {
            setReportData(res.data);
        });
    }

    useEffect(() => {
        getTopologyByTraceId();
    }, [])

    const toggleTopologyFullScreen = () => {
        setTopologyFullScreen(!topologyFullScreen);
    }

    return (
        <div className='root_cause_warp'> 
            <Spin spinning={loading}>
                <div className='top_info_warp'>
                    <Descriptions>
                        <Descriptions.Item label="故障描述">{entryTrace.url}的请求时间异常</Descriptions.Item>
                        <Descriptions.Item label="故障发生时间">{ formatUnit(entryTrace.startTime / 1000000, 'date') }</Descriptions.Item>
                        <Descriptions.Item label="请求耗时">{ formatUnit(entryTrace.totalTime, 'ns') }</Descriptions.Item>
                    </Descriptions>
                    <div className='duration'>
                        <TimeProgress p90={entryTrace.p90? entryTrace.p90 : 0} data={entryTrace.totalTime ? entryTrace.totalTime : 0}/>
                    </div>
                </div>
                <Modal destroyOnClose visible={topoModalVisible} title="全链路调用火焰图" onCancel={() => setTopoModalVisible(false)} footer={null} width="90vw">
                    <TraceTopology data={topologyTreeData}/>
                </Modal>
                <div className='cause_content'>
                    <Row gutter={16} style={{ height: '100%' }}>
                        <Col span={6} style={{ height: '100%' }}>
                            <div id='left_chart_warp' className={`callChart_warp ${topologyFullScreen ? 'full-screen' : ''}`}>
                                <div className='btn_info'>
                                    <Button onClick={() => setTopoModalVisible(true)} size='middle'>全链路调用</Button>
                                </div>
                                {/* <CallChart data={topologyData}/> */}
                                <SingleLinkTopology data={topologyData}/>
                            </div>
                        </Col>
                        <Col span={18} style={{ height: '100%' }}>
                            <div className='cause_steps'>
                                <div className='flex-jc-sb f-mb10'>
                                    <h3 style={{ margin: 0 }}>根因分析</h3>
                                    <Space>
                                        <Button size="middle">关闭故障</Button>
                                        <Button size="middle">详细数据</Button>
                                        <Button size="middle">全屏</Button>
                                    </Space>
                                </div>
                                <div className='report_result'>
                                    <span>故障根因：</span>
                                    <span>{ reportData.conclusion?.message }</span>
                                </div>
                                {
                                    reportData?.reports ? reportData.reports.map((item, idx) => <CauseStep key={idx} data={item} showArrow={idx !== reportData.reports.length - 1}/>) : <Empty></Empty>
                                }
                            </div>
                        </Col>
                    </Row>
                </div>
            </Spin>
        </div>
    );
}

export default RootCause;
