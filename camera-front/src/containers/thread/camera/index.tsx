import React from 'react';
import { Button, Switch, Empty, Tooltip } from 'antd';
import { GatewayOutlined, RollbackOutlined, QuestionCircleOutlined } from '@ant-design/icons';
import tooltipPng from '@/assets/images/tooltip.png';
import _ from 'lodash';
import './index.less';
import { IOption } from './types'; 
import Camera from './draw2';

interface IProps {
    option: IOption;
    style?: React.CSSProperties;
}
interface IState {
    supportAddLine: boolean;
    supportTrace: boolean;
    supportBrush: boolean;
    showJavaLock: boolean;
    showLog: boolean;
}
class CameraWarp extends React.Component<IProps, IState> {
    constructor(props: IProps) {
        super(props);
        this.state = {
            supportAddLine: false,
            supportBrush: false,
            supportTrace: true,
            showJavaLock: true,
            showLog: true
        }
    }
    camera = new Camera(this.props.option);
    observer: any = null;
    
    componentDidMount() {
        this.print();
        this.addMutationObserver();
    }
    componentDidUpdate(prevProps: Readonly<IProps>): void {
        if (!_.isEqual(prevProps.option.data, this.props.option.data)) {
            this.clearSupport();
            this.camera = new Camera(this.props.option);
            this.print();
        }
    }
    componentWillUnmount() {
        this.setState = () => {
            return;
        };
        if (this.observer) {
            this.observer = null;
        }
    }

    addMutationObserver = () => {
        let MutationObserver = window.MutationObserver;
        // || window.WebKitMutationObserver || window.MozMutationObserver;
        let element = document.getElementById('right_thread_warp')
        this.observer = new MutationObserver((mutationList) => {
            this.camera.changeSize();
        });
        this.observer.observe(element, { attributes: true, attributeFilter: ['style', 'class'], attributeOldValue: true })
    }

    clearSupport = () => {
        this.setState({
            supportAddLine: false,
            supportBrush: false,
            supportTrace: true,
            showJavaLock: true,
            showLog: true
        });
    }
    print = () => {
        const {option} = this.props;
        console.log('print', option.data);
        if (option.data && option.data.length > 0) {
            this.camera.draw();
        }
    }

    toggleSupportAddLine = () => {
        this.setState((prevState) => ({
            supportAddLine: !prevState.supportAddLine
        }));
        this.state.supportAddLine ? this.camera.removeAddLine() : this.camera.supportAddLine();
    }
    traceAnaliysis = () => {
        if (this.props.option.data.length > 0) {
            this.setState((prevState) => ({
                supportTrace: !prevState.supportTrace
            }), () => {
                this.camera.startTrace(this.state.supportTrace);
            })
        }
    }
    toggleChartBrush = () => {
        this.camera.addChartBrush();
        // this.setState((prevState) => ({
        //     supportBrush: !prevState.supportBrush
        // }));
        // this.state.supportBrush ? this.camera.removeChartBrush() : this.camera.addChartBrush();
    }
    resetChartBrush = () => {
        this.camera.resetChartBrush()
    }
    toggleJavaLockBtn = () => {
        this.setState((prevState) => ({
            showJavaLock: !prevState.showJavaLock
        }));
        this.state.showJavaLock ? this.camera.hideJavaLock() : this.camera.showJavaLock();
    }
    toggleLogBtn = () => {
        this.setState((prevState) => ({
            showLog: !prevState.showLog
        }));
        this.state.showLog ? this.camera.hideLog() : this.camera.showLog();
    }

    render() {
        const {option} = this.props;
        const {supportAddLine, supportBrush, showJavaLock, showLog, supportTrace} = this.state;
        return (
            <div id="camera_chart_warp">
                <div className='header'>
                    <div>
                        <Button size="small" onClick={this.toggleSupportAddLine} style={{ marginRight: 10 }}>{supportAddLine ? '取消关键时刻' : '添加关键时刻'}</Button>
                        <Button size="small" onClick={this.traceAnaliysis} type={supportTrace ? 'primary' : 'default'} ghost={supportTrace}>Trace分析</Button>
                    </div>
                    <div className='toggle_print'>
                        <Tooltip title="区域缩放">
                            <GatewayOutlined className={`operate_icon ${supportBrush && 'active'}`} onClick={this.toggleChartBrush}/>
                        </Tooltip>
                        <Tooltip title="区域缩放还原">
                            <RollbackOutlined className='operate_icon' onClick={this.resetChartBrush}/>
                        </Tooltip>
                        <span>Java Lock</span>
                        <Switch checked={showJavaLock} onChange={this.toggleJavaLockBtn} size='small'/>
                        <span style={{ marginLeft: 10 }}>Log事件</span>
                        <Switch checked={showLog} onChange={this.toggleLogBtn} size='small'/>
                        <div className='operate_tooltip'>
                            <QuestionCircleOutlined className='operate_icon'/>
                            <img alt='图例' src={tooltipPng}/>
                        </div>  
                    </div>
                </div>
                <div id='camera' className='camera_charts'>
                    {
                        (option.data && option.data.length > 0) ? <>
                            {/* <div className='top_time_warp'> */}
                                <svg id='top_time_svg'></svg>
                            {/* </div> */}
                            <div className='main_chart'>
                                <svg id='camera_svg'></svg>
                            </div>
                            <div className='bottom_xaxis_warp'>
                                <svg id='bottom_xaxis_svg'></svg>
                            </div>
                            <div id='tooltip_warp' className='camera_tooltip'></div>
                        </> : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} />
                    }
                </div>
            </div>
        )
    }
    
}

// function dataEqual(prevProps, nextProps) {
//     return _.isEqual(prevProps.option.data, nextProps.option.data);
// }
// export default React.memo(forwardRef(CameraWarp), dataEqual);
export default CameraWarp;