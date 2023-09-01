import React, { useEffect, useState } from 'react';
import { Row, Col, Button, Space, Spin, Descriptions, Empty } from 'antd';
import _ from 'lodash';
import './index.less';
import TimeProgress from './components/timeProgress';
import CallChart from './components/callChart';
import SingleLinkTopology from './components/svgTopology';
import CauseStep from './components/causeStep';

import { formatUnit } from '@/services/util';
import { useSearchParams } from '@/services/hooks';
import { getTraceTopology, getCauseReports } from '@/request';


function RootCause() {
    const { traceId } = useSearchParams();
    const [loading, setLoading] = useState<boolean>(false);
    const [topologyFullScreen, setTopologyFullScreen] = useState<boolean>(false);
    const [causeFullScreen, setCauseFullScreen] = useState<boolean>(false);
    const [entryTrace, setEntryTrace] = useState<any>({});
    const [topologyData, setTopologyData] = useState({});
    const [reportData, setReportData] = useState<any>({});

    const getTopologyByTraceId = () => {
        setLoading(true);
        getTraceTopology(traceId).then(res => {
            let { entry, nodes, edges } = res.data.data;
            let spanId = '';
            _.forEach(nodes, node => {
                _.forEach(node.list, opt => {
                    if (opt.isMutated && opt.isProfiled) {
                        node.is_warn = true;
                        spanId = opt.spanId;
                    }
                })
            });
            setEntryTrace(entry);
            setTopologyData({ nodes, edges });

            getReportsData(spanId);
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
                        <Descriptions.Item label="故障描述">{entryTrace.contentKey}的请求时间异常</Descriptions.Item>
                        <Descriptions.Item label="故障发生时间">{ formatUnit(entryTrace.startTime / 1000000, 'date') }</Descriptions.Item>
                        <Descriptions.Item label="请求耗时">{ formatUnit(entryTrace.totalTime, 'ns') }</Descriptions.Item>
                    </Descriptions>
                    <div className='duration'>
                        <TimeProgress p90={entryTrace.p90? entryTrace.p90 : 0} data={entryTrace.totalTime ? entryTrace.totalTime : 0}/>
                    </div>
                </div>
                <div className='cause_content'>
                    <Row gutter={16} style={{ height: '100%' }}>
                        <Col span={6} style={{ height: '100%' }}>
                            <div id='left_chart_warp' className={`callChart_warp ${topologyFullScreen ? 'full-screen' : ''}`}>
                                {/* <div className='btn_info'>
                                    <Button onClick={toggleTopologyFullScreen} size='middle'>全屏</Button>
                                </div> */}
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
