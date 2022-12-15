import React, { useEffect, useState, useRef } from 'react';
import { useSearchParams  } from "react-router-dom";
import { Button, Input, Tag, Checkbox, Select, AutoComplete, message, Spin, Progress, TreeSelect, Tooltip} from 'antd';
import { SyncOutlined, MenuFoldOutlined, MenuUnfoldOutlined, DownOutlined, UpOutlined, ArrowLeftOutlined, ArrowRightOutlined, FilterOutlined, FullscreenOutlined, FullscreenExitOutlined, ExclamationCircleOutlined } from '@ant-design/icons';
import _ from 'lodash';
import CustomDatePicker from '@/containers/components/customDatePicker';
import CameraWarp from './camera';
import './index.less';

import { protoclList, eventList as staticEventList, netReadTypes, netWriteTypes, dataHandle, buildFilterList, formatTimeNs, getLineTimesList } from './camera/util';
import { getTraceList, getTraceData, getTracePayload, getFolderList, getFileList, getFileDetail } from '@/request';
import { IFilterField } from './type';
import Flamegraph from './Flamegraph';
import EventDetail from './EventDetail';
import { IOption, IFilterParams, IEvent, ILineTime } from './camera/types';
import { setStore, getStore, defaultTimeInfo, translateRecentTime } from '@/services/util';
import moment from 'moment';

function Thread() {
    const [searchParams] = useSearchParams();
    const [showESQuery, setShowESQuery] = useState(false);
    // es的接口查询相关数据需要的参数
    const [pid, setPid] = useState(getStore('pid') || '');
    const [testStartTimestamp, setStartTimeStamp] = useState<number | null>(parseInt(getStore('startTimestamp') as string) || null);
    const [testEndTimestamp, setEndTimestamp] = useState<number | null>(parseInt(getStore('endTimestamp') as string) || null);
    const [testProtocl, setTestProtocl] = useState('http');
    // node层直接查询文件的相关参数
    const [folderList, setFolderList] = useState([]);
    const [folderName, setFolderName] = useState('');
    const [traceFileList, setTraceFileList] = useState([]);
    const [urlTraceFileList, setUrlTraceFileList] = useState({});
    const [traceUrlList, setTraceUrlList] = useState<string[]>([]);
    const [traceUrl, setTraceUrl] = useState<string>('');
    const [testFileName, setTestFileName] = useState('');
    
    const [tiemValue, setTimeValue] = useState();
    const [traceList, setTraceList] = useState<any[]>([]);
    const [nowTrace, setNowTrace] = useState<any>({});
    const [nowTraceIdx, setNowTraceIdx] = useState<number>(0);
    const [nowTraceData, setNowTraceData] = useState<any>([]);
    const [requestEventInfo, setRequestEventInfo] = useState<any>([]);
    const [nowEvent, setNowEvent] = useState<any>({});
    const [traceId, setTraceId] = useState('');     // 返回的trace.labels中没有traceId需要去cpuEvents里面匹配一次，若存在需要回填traceId

    const [lineTimeList, setLineTimeList] = useState<ILineTime[]>([]);
    const [loading, setloading] = useState<boolean>(false);
    const [infoloading, setInfoloading] = useState<boolean>(false);

    const [nowTraceTimeRange, setNowTraceTimeRange] = useState<Date[]>([]);
    const [threadList, setThreadList] = useState<any>([]);
    const [selectThreadList, setSelectThreadList] = useState<any[]>([]);
    const [chartTid, setChartTid] = useState<number | null>(null);
    const [logValue, setLogValue] = useState<string>('');
    const [selectLogList, setSelectLogList] = useState<any[]>([]);
    const [fileList, setFileList] = useState<any[]>([]);
    const [fileValue, setFileValue] = useState<string>('');
    const [selectFileList, setSelectFileList] = useState<any[]>([]);

    const [eventList, setEventList] = useState<IEvent[]>([]);
    const [selectEventList, setSelectEventList] = useState<any>([]);

    const [showFullScreen, setShowFullScreen] = useState<boolean>(false);
    const [hideFilterWarp, setHideFilterWarp] = useState<boolean>(false);
    const [hideFilterFieldList, setHideFilterFieldList] = useState<string[]>([]);
    const [showEventDetail, setShowEventDetail] = useState<boolean>(true);

    let cameraRef = useRef();

    const getTraceDetail = (trace: any) => {
        let traceStartTimestamp = Math.ceil(trace.timestamp / 1000000);
        let totalTime = _.find(trace.metrics, {Name: 'request_total_time'});
        let traceEndTimestamp = Math.ceil((trace.timestamp + totalTime.Data.Value) / 1000000);
        const requestTimes = [traceStartTimestamp, traceEndTimestamp];
        let startTimestamp = traceStartTimestamp - 5 * 1000;
        let endTimestamp = traceEndTimestamp + 5 * 1000;
        const params = {
            pid: trace.labels.pid,
            startTimestamp: startTimestamp,
            endTimestamp: endTimestamp
        }
        getTraceData(params).then(res => {
            setloading(false);
            if (res.data.length > 0) {
                let result = _.map(res.data, 'labels');
                _.forEach(result, item => {
                    item.cpuEvents = JSON.parse(item.cpuEvents as string);
                    item.javaFutexEvents = JSON.parse(item.javaFutexEvents as string);
                    item.transactionIds = JSON.parse(item.transactionIds as string);
                });
                const { data, eventlist, traceTimes, requestInfo, traceId } = dataHandle(result, [startTimestamp, endTimestamp], trace);
                const { threadList, fileList } = buildFilterList(data);

                const linetimes = getLineTimesList(requestTimes, traceTimes);
                console.log(data, linetimes);
                setNowTraceData(data);
                setNowTraceTimeRange([new Date(startTimestamp), new Date(endTimestamp)]);
                setLineTimeList(linetimes);
                setThreadList(threadList);
                setFileList(fileList);
                setEventList(eventlist);
                setRequestEventInfo(requestInfo);
                setTraceId(traceId);
            } else {
                setNowTraceData([]);
                setThreadList([]);
                setFileList([]);
                setRequestEventInfo([]);
                setEventList(staticEventList);
                setNowEvent({});
            }
        });
    }
    const initData = (time) => {
        setloading(true);
        const params: any = {
            startTimestamp: time.from,
            endTimestamp: time.to,
            isServer: true,
            protocol: testProtocl
        }
        if (pid) {
            params.pid = pid;
        }
        if (testStartTimestamp && testStartTimestamp > 0 && testEndTimestamp && testEndTimestamp > 0) {
            params.startTimestamp = testStartTimestamp;
            params.endTimestamp = testEndTimestamp;
        }
        getTraceList(params).then(res => {
            let result: any = [];
            if (pid) {
                result = _.filter(res.data, item => item.labels.pid === parseInt(pid));
            } else {
                result = res.data;
            }
            if (result.length > 0) {
                setTraceList(result);
                setNowTrace(result[0]);
                if (result[0].labels?.trace_id) {
                    setTraceId(result[0].labels.trace_id);
                }
                setNowTraceIdx(0);
                getTraceDetail(result[0]);
            } else {
                setloading(false);
                message.warning('暂无数据');
                setTraceList([]);
                setNowTrace({});
            }
        });
    }
    const init2 = () => {
        let time;
        if (getStore('time')) {
            let times = JSON.parse(getStore('time') as string);
            time = translateRecentTime(times);
        } else {
            time = defaultTimeInfo;
        }
        setTimeValue(time);
        const timeRange = {
            from: new Date(time.from).getTime(),
            to: new Date(time.to).getTime()
        }
        initData(timeRange);
    }
    // node 接口直接读取文件列表
    const getAllFolderList = () => {
        getFolderList().then(res => {
            if (res.data.success) {
                setFolderList(res.data.data);
            } else {
                setFolderList([]);
            }
        });
    }
    const getAllTraceFile = (folder) => {
        setFolderName(folder);
        getFileList({folderName: folder}).then(res => {
            if (res.data.success && res.data.data.length > 0) {
                // let urlList: string[] = _.uniq(_.map(res.data.data, 'contentKey'));
                let urlFileList = _.groupBy(res.data.data, 'contentKey');
                let urlList: string[] = _.keys(urlFileList);
                setTraceUrlList(urlList);
                setTraceUrl(urlList[0]);
                // setTraceFileList(res.data.data);
                setUrlTraceFileList(urlFileList);
                let fileList: any = urlFileList[urlList[0]];
                setTraceFileList(fileList);
                selectFileName(fileList[0].fileName, folder);
            } else {
                setTraceFileList([]);
            }
        });
    }
    const selectFolder = (folder) => {
        getAllTraceFile(folder);
    }
    const changeTraceUrl = (url) => {
        setTraceUrl(url);
        let fileList: any = urlTraceFileList[url];
        setTraceFileList(fileList);
        selectFileName(fileList[0].fileName, folderName);
    }
    const selectFileName = (fileName, folder) => {
        setTestFileName(fileName);
        getFileDetail({folderName: folder, fileName}).then(res => {
            if (res.data.success) {
                let {trace: traceData, cpuEvents} = res.data.data;
                setTraceList([traceData]);
                setNowTrace(traceData);
                console.log('traceData', traceData);
                if (traceData.labels?.trace_id) {
                    setTraceId(traceData.labels.trace_id);
                }
                setNowTraceIdx(0);

                let traceStartTimestamp = Math.ceil(traceData.timestamp / 1000000);
                let totalTime = _.find(traceData.metrics, {Name: 'request_total_time'});
                let traceEndTimestamp = Math.ceil((traceData.timestamp + totalTime.Data.Value) / 1000000);
                const requestTimes = [traceStartTimestamp, traceEndTimestamp];
                let startTimestamp = traceStartTimestamp - 5 * 1000;
                let endTimestamp = traceEndTimestamp + 5 * 1000;

                if (cpuEvents.length > 0) {
                    const { data, eventlist, traceTimes, requestInfo, traceId } = dataHandle(cpuEvents, [startTimestamp, endTimestamp], traceData);
                    const { threadList, fileList } = buildFilterList(data);
    
                    const linetimes = getLineTimesList(requestTimes, traceTimes);
                    console.log(data, linetimes);
                    setNowTraceData(data);
                    setNowTraceTimeRange([new Date(startTimestamp), new Date(endTimestamp)]);
                    setLineTimeList(linetimes);
                    setThreadList(threadList);
                    setFileList(fileList);
                    setEventList(eventlist);
                    setRequestEventInfo(requestInfo);
                    setTraceId(traceId);
                    setNowEvent({});
                } else {
                    setNowTraceData([]);
                    setThreadList([]);
                    setFileList([]);
                    setEventList(staticEventList);
                    setRequestEventInfo([]);
                    setNowEvent({});
                }
            } else {
                message.error(res.data.message);
            }
        });
    }

    useEffect(() => {
        setShowESQuery((searchParams.get('query') && searchParams.get('query') === 'es') ? true : false);
        let theme: any = searchParams.get('theme') || 'light';
        let body = document.getElementsByTagName('body')[0];
        body.className = `${theme}-theme`;
        setStore('theme', theme);
    }, [searchParams]);

    useEffect(() => {
        getAllFolderList();
    }, []);

    const timeChange = (time) => {
        setStore('time', JSON.stringify(time));
        const timeRange = {
            from: new Date(time.from).getTime(),
            to: new Date(time.to).getTime()
        }
        initData(timeRange);
    }

    const toggleFilterWarp = () => {
        setHideFilterWarp(!hideFilterWarp);
    }
    const toggleFilterField = (field: IFilterField) => {
        if (hideFilterFieldList.indexOf(field) > -1) {
            let list = [...hideFilterFieldList];
            list.splice(hideFilterFieldList.indexOf(field), 1);
            setHideFilterFieldList(list);
        } else {
            setHideFilterFieldList([...hideFilterFieldList, field])
        }
    }

    const toggleTrace = (type: 'pre' | 'next') => {
        let idx = type === 'pre' ? nowTraceIdx - 1 : nowTraceIdx + 1;
        setNowTraceIdx(idx);
        setNowTrace(traceList[idx]);
        setloading(true);
        getTraceDetail(traceList[idx]);
        resetAllField();
    }

    const fullScreen = () => {
        setShowFullScreen(!showFullScreen);
        let camera = (cameraRef.current as any)?.camera;
        camera.changeSize()
    }

    // 线程搜索
    const selectThread = (value) => {
        if (_.findIndex(selectThreadList, { value: value.value }) === -1) {
            let list = [...selectThreadList, { label: value.label, value: value.value }];
            setSelectThreadList(list);
        }
    }
    const removeSelectThread = (list) => {
        let result = _.filter(selectThreadList, item => list.indexOf(item.value) > -1);
        setSelectThreadList(result);
    }
    // 日志
    const logKeyDown = (e) => {
        if (e.keyCode === 13) {
            let value = e.target.value;
            if (selectLogList.indexOf(value) === -1) {
                setSelectLogList([...selectLogList, value]);
                setLogValue('');
            }
        }
    }
    const removeLog = (list) => {
        setSelectLogList(list);
    }
    // 操作文件搜索
    const addSelectFile = (value) => {
        if (selectFileList.indexOf(value) === -1) {
            setSelectFileList([...selectFileList, value]);
            setFileValue('');
        }
    }
    const fileKeyDown = (e) => {
        if (e.keyCode === 13) {
            addSelectFile(e.target.value);
        }
    }
    const removeSelectFile = (list) => {
        setSelectFileList(list);
    }
    // 事件类型筛选
    const selectShowEvent = (eventNames: any[]) => {
        setSelectEventList(eventNames);
    }
    const reprintByFilter = () => {
        const filterParams: IFilterParams = {
            threadList: _.map(selectThreadList, 'value'),
            logList: selectLogList,
            fileList: selectFileList,
            eventList: selectEventList.length > 0 ? selectEventList : [..._.map(eventList, 'value'), 'net', 'file']
        };
        (cameraRef.current as any).closeTraceAnaliysis();
        let camera = (cameraRef.current as any)?.camera;
        camera.reprintByFilter(filterParams);
        camera.closeShowTrace();
    }
    // 清空重置所有的筛选条件
    const resetAllField = () => {
        // TODO 四个清空操作是异步的  -----  这是一个bug
        setSelectThreadList([]);
        setSelectLogList([]);
        setSelectFileList([]);
        setSelectEventList([]);
        // reprintByFilter();
    }

    const clickRequestEvent = (e: any) => {
        let camera = (cameraRef.current as any)?.camera;
        camera.shiningEvent(e.type);
    }
    const handleEventClick = (evt: any) => {
        console.log('event', evt);
        if (showESQuery) {
            if (evt.type === 'net' && evt.info) {
                setInfoloading(true);
                const tempIp = evt.info.file.split('->');
                const [srcIp, srcPort] = tempIp[0].split(':');
                const [dstIp, dstPort] = tempIp[1].split(':');
                let type = '';
                if (netReadTypes.indexOf(evt.info.operate) > -1) {
                    type = parseInt(evt.info.requestType) === 1 ? 'request' : 'response'
                }
                if (netWriteTypes.indexOf(evt.info.operate) > -1) {
                    type = parseInt(evt.info.requestType) === 1 ? 'response' : 'request'
                }
                const params = {
                    pid: nowTrace.labels.pid,
                    type: type,
                    startTimestamp: parseInt(evt.endTime) - 2,
                    endTimestamp: parseInt(evt.endTime) + 2,
                    srcIp, 
                    srcPort, 
                    dstIp, 
                    dstPort
                };
                getTracePayload(params).then(res => {
                    const event: any = {
                        ...evt,
                        message: type === 'request' ? res.data.labels?.request_payload : res.data.labels?.response_payload,
                        requestType: type,
                        traceInfo: res.data
                    };
                    console.log('net event: ', event);
                    setNowEvent(event);
                    setTimeout(() => {
                        setInfoloading(false);
                    }, 100);
                });
            } else {
                if (evt.active) {
                    evt.message = evt.eventType === "netread" ? nowTrace.labels.request_payload : nowTrace.labels.response_payload;
                }
                setNowEvent(evt);
            }
        } else {
            if (evt.active) {
                evt.message = evt.eventType === "netread" ? nowTrace.labels.request_payload : nowTrace.labels.response_payload;
            }
            setNowEvent(evt);
        }
    }

    useEffect(() => {
        if (chartTid) {
            let thread = _.find(threadList, {tid: chartTid});
            if (_.findIndex(selectThreadList, { value: chartTid }) === -1) {
                let list = [...selectThreadList, ...[{ label: thread.name, value: thread.tid }]];
                setSelectThreadList(list);
            }
        }
    }, [chartTid]);

    const handleNameClick = (tid: string) => {
        setChartTid(parseInt(tid));
    }

    const toggleEventDetail = () => {
        setShowEventDetail(!showEventDetail);
    }

    const changePid = (e) => {
        setPid(e.target.value);
        setStore('pid', e.target.value);
    }
    const changeStartTimestamp = (e) => {
        setStartTimeStamp(parseInt(e.target.value))
        setStore('startTimestamp', e.target.value);
    }
    const changeEndTimestamp = (e) => {
        setEndTimestamp(parseInt(e.target.value))
        setStore('endTimestamp', e.target.value);
    }
    const option: IOption = {
        data: nowTraceData,
        trace: nowTrace,
        traceId: nowTrace?.labels?.trace_id || '',
        lineTimeList: lineTimeList,
        timeRange: nowTraceTimeRange,
        nameClick: handleNameClick,
        eventClick: handleEventClick
    };
    return (
        <div className='thread_warp'>
            <header className='thread_header'>
                <div className='thread_header_text'>
                    <span>应用运行分析</span>
                    {
                        showESQuery ? <React.Fragment>
                            <Input type="number" style={{ width: 100 }} value={pid} onChange={changePid} />
                            <Input type="number" style={{ width: 180 }} value={testStartTimestamp} onChange={changeStartTimestamp} />
                            <Input type="number" style={{ width: 180 }} value={testEndTimestamp} onChange={changeEndTimestamp} />
                            <Select style={{ width: 100 }} value={testProtocl} onChange={v => setTestProtocl(v)}>
                                {
                                    protoclList.map(item => <Select.Option key={item.value} value={item.value}>{item.field}</Select.Option>)
                                }
                            </Select>
                            <Button onClick={init2}>查询</Button>   
                        </React.Fragment> : <React.Fragment>
                            <span className='small_title' style={{ marginLeft: 10 }}>Container：</span>
                            <TreeSelect style={{ width: 180, marginRight: 10 }} dropdownStyle={{ maxHeight: 400, overflow: 'auto' }} dropdownMatchSelectWidth={280} value={folderName} treeData={folderList} treeDefaultExpandAll onChange={selectFolder}/>
                            <span className='small_title'>Trace：</span>
                            <Select value={traceUrl} style={{ width: 120, marginRight: 10 }} dropdownMatchSelectWidth={240} onChange={changeTraceUrl}>
                                {
                                    traceUrlList.map((url: string, idx) => <Select.Option key={idx} value={url}>{url}</Select.Option>)
                                }
                            </Select>
                            <span className='small_title'>Profile：</span>
                            <Select style={{ width: 120 }} dropdownMatchSelectWidth={240} value={testFileName} onChange={(v) => selectFileName(v, folderName)}>
                                {
                                    traceFileList.map((item: any, idx) => <Select.Option key={idx} value={item.fileName}>{item.showFileName}</Select.Option>)
                                }
                            </Select>
                        </React.Fragment>
                    }         
                </div>
                <div className='thread_header_operation'>
                    {
                        showESQuery ? <>
                            <CustomDatePicker timeInfo={tiemValue} onChange={timeChange}/>
                            <Button onClick={() => window.location.reload()} icon={<SyncOutlined />} className='f-ml8' />
                        </> : null
                    }
                </div>
            </header>
            <div className='thead_body'>
                {
                    hideFilterWarp && <Button className='toggle_btn' shape='circle' size='small' icon={<MenuUnfoldOutlined />} onClick={toggleFilterWarp} />
                }
                <div className='filter_panel' style={{ display: hideFilterWarp ? 'none' : 'block' }}>
                    <Button className='toggle_btn' shape='circle' size='small' icon={<MenuFoldOutlined />} onClick={toggleFilterWarp} />
                    <div className='filter_body'>
                        <div className='header_title'>
                            <span>线程筛选</span>
                        </div>
                        <div className='filter_field_warp'>
                            <div className='field_header'>
                                <div>
                                    <span>线程搜索</span>
                                    <FilterOutlined className={`operate_icon ${selectThreadList.length > 0 && 'active'}`} onClick={() => setSelectThreadList([])} />
                                </div>
                                <div onClick={() => toggleFilterField('thread')} className='operate_icon'>
                                    {
                                        hideFilterFieldList.indexOf('thread') === -1 ? <DownOutlined /> : <UpOutlined />
                                    }
                                </div>
                            </div>
                            {
                                hideFilterFieldList.indexOf('thread') === -1 ? <div className='field_body'>
                                    <Select value={{ label: '', value: '' }} style={{ width: '100%' }} onChange={selectThread} labelInValue
                                        showSearch filterOption={(input, option: any) =>
                                            option.props.children.toLowerCase().indexOf(input.toLowerCase()) >= 0
                                        }>
                                        {
                                            threadList.map(item => <Select.Option key={item.tid} value={item.tid}>{item.name}</Select.Option>)
                                        }
                                    </Select>
                                    {
                                        selectThreadList.length > 0 ? <ul className='event_legend_list'>
                                            <Checkbox.Group value={_.map(selectThreadList, 'value')} onChange={removeSelectThread}>
                                                {
                                                    selectThreadList.map(opt => <li key={opt.value}><Checkbox value={opt.value} className='checklabel'>{opt.label}</Checkbox></li>)
                                                }
                                            </Checkbox.Group>
                                        </ul> : null
                                    }
                                </div> : null
                            }
                        </div>
                        <div className='filter_field_warp'>
                            <div className='field_header'>
                                <div>
                                    <span>日志搜索</span>
                                </div>
                                <div onClick={() => toggleFilterField('log')} className='operate_icon'>
                                    {
                                        hideFilterFieldList.indexOf('log') === -1 ? <DownOutlined /> : <UpOutlined />
                                    }
                                </div>
                            </div>
                            {
                                hideFilterFieldList.indexOf('log') === -1 ? <div className='field_body'>
                                    <Input value={logValue} onChange={(e) => setLogValue(e.target.value)} onKeyDown={logKeyDown}/>
                                    {
                                        selectLogList.length > 0 ? <ul className='event_legend_list'>
                                            <Checkbox.Group value={selectLogList} onChange={removeLog}>
                                                {
                                                    selectLogList.map(log => <li key={log}><Checkbox value={log} className='checklabel'>{log}</Checkbox></li>)
                                                }
                                            </Checkbox.Group>
                                        </ul> : null
                                    }
                                </div> : null
                            }
                        </div>
                        <div className='filter_field_warp'>
                            <div className='field_header'>
                                <div>
                                    <span>操作文件搜索</span>
                                </div>
                                <div onClick={() => toggleFilterField('file')} className='operate_icon'>
                                    {
                                        hideFilterFieldList.indexOf('file') === -1 ? <DownOutlined /> : <UpOutlined />
                                    }
                                </div>
                            </div>
                            {
                                hideFilterFieldList.indexOf('file') === -1 ? <div className='field_body'>
                                    <AutoComplete value={fileValue} options={fileList} style={{ width: '100%' }} defaultOpen={false} onSelect={addSelectFile} filterOption={(input, option: any) =>
                                        option.value.toLowerCase().indexOf(input.toLowerCase()) >= 0
                                    } onKeyDown={fileKeyDown} onChange={v => setFileValue(v)}/>
                                    {
                                        selectFileList.length > 0 ? <ul className='event_legend_list'>
                                            <Checkbox.Group value={selectFileList} onChange={removeSelectFile}>
                                                {
                                                    selectFileList.map(file => <li key={file}><Checkbox value={file} className='checklabel'>{file}</Checkbox></li>)
                                                }
                                            </Checkbox.Group>
                                        </ul> : null
                                    }
                                </div> : null
                            }
                        </div>
                        <div className='filter_field_warp'>
                            <div className='field_header'>
                                <div>
                                    <span>事件类型</span>
                                    <FilterOutlined className={`operate_icon ${selectEventList.length > 0 && 'active'}`} onClick={() => setSelectEventList([])} />
                                </div>
                                <div onClick={() => toggleFilterField('event')} className='operate_icon'>
                                    {
                                        hideFilterFieldList.indexOf('event') === -1 ? <DownOutlined /> : <UpOutlined />
                                    }
                                </div>
                            </div>
                            {
                                hideFilterFieldList.indexOf('event') === -1 ? <div className='field_body' style={{ padding: '0 0 10px 0' }}>
                                    <ul className='event_legend_list'>
                                        <Checkbox.Group value={selectEventList} onChange={selectShowEvent}>
                                            {
                                                eventList.map(opt => <li key={opt.value}><Checkbox value={opt.value}>
                                                    <span>{opt.name}</span>
                                                    {
                                                        _.isNumber(opt.count) && <span className='count_tag' style={{ backgroundColor: opt.color }}>{opt.count}</span>
                                                    }
                                                </Checkbox></li>)
                                            }
                                        </Checkbox.Group>
                                    </ul>
                                </div> : null
                            }
                        </div>
                    </div>
                    <div className='footer_btn'>
                        <Button type='primary' style={{ marginRight: 10 }} onClick={reprintByFilter}>查询</Button>
                        <Button onClick={resetAllField}>重置</Button>
                    </div>
                </div>
                <div className={`thead_content ${showFullScreen && 'fullScreen'}`} id='right_thread_warp' style={{ width: hideFilterWarp ? '100%' : 'calc(100% - 320px)' }}>
                    <div className='url_warp'>
                        <div className='url_info'>
                            <h3>{nowTrace?.labels?.http_method}：{nowTrace?.labels?.http_url}</h3>
                            <div>
                                <Tag>Trace ID：{nowTrace?.labels ? (traceId ? <span>{traceId}</span> : <Tooltip title="请先安装Skywalking探针，否则无法标注关键线程">
                                    <ExclamationCircleOutlined className='shinning_icon'/>
                                </Tooltip>) : null}</Tag>
                                <Tag>协议类型：{nowTrace?.labels?.protocol}</Tag>
                                <Tag>响应时间：{formatTimeNs(_.find(nowTrace?.metrics, {Name: 'request_total_time'})?.Data.Value || 0)}</Tag>
                                <Tag>返回码：{nowTrace?.labels?.http_status_code}</Tag>
                                <Tag>TimeStamp：{nowTrace?.timestamp ? moment(Math.floor(nowTrace?.timestamp / 1000000)).format('YYYY-MM-DD HH:mm:ss') : ''}</Tag>
                            </div>
                        </div>
                        <div>
                            {
                                showESQuery ? <React.Fragment>
                                    <Button icon={<ArrowLeftOutlined />} className='f-ml8' onClick={() => toggleTrace('pre')} disabled={nowTraceIdx === 0 || traceList.length === 0} />
                                    <Button icon={<ArrowRightOutlined />} className='f-ml8' onClick={() => toggleTrace('next')} disabled={nowTraceIdx === (traceList.length - 1) || traceList.length === 0} />
                                </React.Fragment> : null
                            }
                            <Button icon={showFullScreen ? <FullscreenExitOutlined/> : <FullscreenOutlined/>} className='f-ml8' onClick={fullScreen}/>
                        </div>
                    </div>
                    {
                        requestEventInfo.length > 0 && <div className='process_over_warp'>
                            <div className='request_info_process_warp'>
                                { 
                                    requestEventInfo.map(item => 
                                        <div className='process_inof' key={item.type} onClick={() => clickRequestEvent(item)}>
                                            <div className='info_warp'>
                                                <div className='event_name'>
                                                    <span>{item.type}</span>
                                                </div>
                                                <span className='event_time'>{item.time}ms</span>
                                            </div>
                                            <div className='process_warp'>
                                                <Progress type="circle" percent={item.timeRate} width={60} trailColor={(getStore('theme') || 'light') === 'light' ? '#3790FF12' : '#696969'} strokeColor='#3790FF' showInfo={false}/>
                                                <div className='process_center_text'>
                                                    <span>{item.timeRate}%</span>
                                                </div>
                                            </div>
                                            {/* <Divider type="vertical" /> */}
                                        </div>
                                    )
                                }
                            </div>
                        </div>
                    }
                    <div className='charts' style={{ height: requestEventInfo.length > 0 ? 'calc(60% - 90px)' : '60%' }}>
                        {
                            loading && <Spin spinning={loading} className="camera_loading"></Spin>
                        }
                        <CameraWarp ref={cameraRef} option={option} />
                    </div>
                    <div className='event_info_warp'>
                        <div className='event_header'>
                            <span>事件详情</span>
                            <div onClick={toggleEventDetail}>
                                {showEventDetail ? <UpOutlined /> : <DownOutlined />}
                            </div>
                        </div>
                        <div className={`event_body ${showEventDetail ? 'show' : 'hide'}`}>
                            <Spin spinning={infoloading}>
                                {nowEvent?.eventType === 'on' ? <Flamegraph data={nowEvent} /> : <EventDetail data={nowEvent} showESQuery={showESQuery} />}
                            </Spin>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    )
}

export default Thread;