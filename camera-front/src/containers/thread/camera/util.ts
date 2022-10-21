import _ from 'lodash';
import * as d3 from 'd3';
import { IEvent, IEventTime, IJavaLock, ILogEvent, IThread, ILineTime } from './types';

export const protoclList = [
    {
        field: 'http',
        value: 'http'
    }, {
        field: 'mysql',
        value: 'mysql'
    }, {
        field: 'dns',
        value: 'dns'
    }, {
        field: 'kafka',
        value: 'kafka'
    }, {
        field: 'redis',
        value: 'redis'
    }, {
        field: 'dubbo',
        value: 'dubbo'
    }, {
        field: 'NOSUPPORT',
        value: 'NOSUPPORT'
    }
];
export const eventList: IEvent[] = [
    {
        name: 'oncpu',
        value: 'on',
        type: 'on',
        fillColor: '#FDE6E4',
        activeColor: '#FADDDB',
        color: '#E10000'
    }, {
        name: 'file-open',
        value: 'fileopen',
        type: 'file',
        fillColor: '#EDE4FF',
        activeColor: '#E5DAFB',
        color: '#A70CC9'
    }, {
        name: 'file-close',
        value: 'fileclose',
        type: 'file',
        fillColor: '#EDE4FF',
        activeColor: '#E5DAFB',
        color: '#A70CC9'
    }, {
        name: 'net-read',
        value: 'netread',
        type: 'net',
        fillColor: '#FEF3E6',
        activeColor: '#F8E8D7',
        color: '#E97A00'
    }, {
        name: 'net-write',
        value: 'netwrite',
        type: 'net',
        fillColor: '#FEF3E6',
        activeColor: '#F8E8D7',
        color: '#E97A00'
    }, {
        name: 'futex',
        value: 'futex',
        type: 'futex',
        fillColor: '#E6F4F5',
        activeColor: '#CDF1F3',
        color: '#0094A5'
    }, {
        name: 'idle',
        value: 'idle',
        type: 'idle',
        fillColor: '#EDF8FF',
        activeColor: '#E1EDF5',
        color: '#0877CB'
    }, {
        name: 'epoll',
        value: 'epoll',
        type: 'epoll',
        fillColor: '#e5f5ea',
        activeColor: '#d0f1da',
        color: '#06a838'
    }, {
        name: 'other',
        value: 'other',
        type: 'other',
        fillColor: '#E5E5E5',
        activeColor: '#DAD8D8',
        color: '#333333'
    }, {
        name: 'MonitorEnter(Java Lock)',
        value: 'MonitorEnter',
        type: 'lock',
        fillColor: '#fffde280',
        activeColor: '#EFEDCA',
        color: '#c0b81c'
    }, {
        name: 'MonitorWait(Java Lock)',
        value: 'MonitorWait',
        type: 'lock',
        fillColor: '#fffde280',
        activeColor: '#EFEDCA',
        color: '#c0b81c'
    }, {
        name: 'UnsafePark(Java Lock)',
        value: 'UnsafePark',
        type: 'lock',
        fillColor: '#fffde280',
        activeColor: '#EFEDCA',
        color: '#c0b81c'
    }
];

export const netReadTypes = ["read", "recvfrom", "recvmsg", "readv", "pread", "preadv"];
export const netWriteTypes = ["write", "sendto", "sendmsg", "writev", "pwrite", "pwritev"];

export const textHandle = (text: string, num: number) => {
    if (text && text.length > num) {
        return text.substring(0, num) + '...';
    } else {
        return text;
    }
}

export const timeFormat = function(time: Date) {
    return d3.timeFormat("%M:%S.%L")(time);
};

export const formatTimeNs = time => {
    let flag = false;
    if (time === 'NaN' || time === undefined) {
        return 0;
    } else {
        flag = time < 0;
        time = Math.abs(time);
        if (time > 3600000000000) {
            time = (time / 3600000000000).toFixed(2) + 'h';
        } else if (time >= 60000000000) {
            time = (time / 60000000000).toFixed(2) + 'm';
        } else if (time >= 1000000000) {
            time = (time / 1000000000).toFixed(2) + 's';
        } else if (time >= 1000000) {
            time = (time / 1000000).toFixed(2) + 'ms';
        } else if (time > 1000) {
            time = (time / 1000).toFixed(2) + 'μs';
        } else {
            time = time.toFixed(2) + 'ns';
        }
        return flag ? '-' + time : time;
    }
};

export const formatTimsToMS = (value: number, int = false) => {
    return int ? parseInt(value / 1000000 + '') : value / 1000000;
}

const eventInfoHandle = (info: string) => {
    let infoList = info.split('@');
    let result: any = {};
    switch(infoList[0]) {
        case 'file':
        case 'net':
            result.type = infoList[0];
            result.operate = infoList[1];
            result.file = infoList[2];
            result.startTime = parseInt(infoList[3]);
            result.duration = parseInt(infoList[4]);
            if (infoList[0] === 'net') {
                result.requestType = infoList[5];
                result.size = infoList[6];
            } else {
                result.size = infoList[5];
            }
            break;
        case 'epoll': 
            result.type = infoList[0];
            result.operate = infoList[1];
            result.file = infoList[2];
            result.size = infoList[3];
            result.timestamp = infoList[4];
            break;
        case 'futex': 
            result.type = infoList[0];
            result.address = infoList[1];
            break;
        default:
            result.type = infoList[0];
            result.address = infoList[1];
            break;
    }
    return result;
}

/**
 * 判断事件的发生结束时间是否在时间区间内
 * @param timeRange 时间区间
 * @param startTime 开始时间
 * @param endTime  结束时间
 * @returns boolean
 */
export const containTime = (timeRange: number[], startTime: number, endTime: number) => {
    let stime = startTime > 10000000000000 ? formatTimsToMS(startTime) : startTime;
    let etime = endTime > 10000000000000 ? formatTimsToMS(endTime) : endTime;
    return etime > timeRange[0] && stime < timeRange[1];
}

/**
 * 针对部分事件，开始时间在时间区间外，结束时间在时间区间内（或者开始在，结束超出），需要对数据的时间进行截取
 * @param startTime 开始时间
 * @param endTime 结束时间
 * @param timeNum 持续时间
 * @param timeRange 时间区间
 * @returns 
 */
const timeHandle = (startTime, endTime, timeNum, timeRange, transform = true) => {
    let time = transform ? formatTimsToMS(timeNum) : timeNum;
    // 若事件的开始事件小于当前请求的时间区间，需要根据当前时间的开始值进行截取，同时重新计算事件耗时
    let stime = transform ? formatTimsToMS(startTime) : startTime;
    if (stime < timeRange[0]) {
        time = time - (timeRange[0] - stime);
        stime = timeRange[0];
    }
    // 若事件的结束时间大于当前请求的时间区间，需要根据当前时间的结束值进行截取，同时重新计算事件耗时
    let etime = transform ? formatTimsToMS(endTime) : endTime;
    if (etime > timeRange[1]) {
        time = time - (etime - timeRange[1]);
        etime = timeRange[1];
    }
    return {stime, etime, time}
}

/**
 * 处理onInfo根offInfo里面的事件详情，各事件用|分割，但是在|分割的事件详情若存在#则表示里面包含多个事件，需要根据#分割，使用info信息里面startTime和duration重新生成事件
 * @param info onInfo根offInfo根据type下标找到的事件string Info
 * @param eventObj 当前根据timeType，typeSpecs和stack 构造的事件详情
 * @param timeRange 当前时间的时间查询范围
 * @returns IEventTime | IEventTime[]
 */
const onOffInfoHandle = (info, eventObj, timeRange) => {
    if (info.indexOf('#') > -1) {
        let result: IEventTime[] = [];
        let infoList = info.split('#');
        _.forEach(infoList, infos => {
            let detailInfo = eventInfoHandle(infos);
            let endTime = detailInfo.startTime + detailInfo.duration;
            if (containTime(timeRange, detailInfo.startTime, endTime)) {
                let { stime, etime, time } = timeHandle(detailInfo.startTime, endTime, detailInfo.duration, timeRange);
                let subEvent: IEventTime = {
                    startTime: stime,
                    endTime: etime,
                    time: time,
                    type: detailInfo.type,
                    stackList: eventObj.stackList
                };
                if (['file', 'net'].indexOf(detailInfo.type) > -1) {
                    subEvent.type = detailInfo.type;
                    subEvent.eventType = detailInfo.type + detailInfo.operate;
                } else {
                    subEvent.type = detailInfo.type;
                    subEvent.eventType = detailInfo.type;
                }
                detailInfo.startTime = formatTimsToMS(detailInfo.startTime);
                detailInfo.duration = formatTimsToMS(detailInfo.duration);
                subEvent.info = detailInfo;
                if (eventObj.endTime > timeRange[0] && eventObj.startTime < timeRange[1]) {
                    result.push(subEvent);
                }
            }
        });
        console.log('带#解析的info结果：', result);
        return result;
    } else {
        let detailInfo = eventInfoHandle(info);
        let result: IEventTime = _.cloneDeep(eventObj);
        if (['file', 'net'].indexOf(detailInfo.type) > -1) {
            let endTime = detailInfo.startTime + detailInfo.duration;
            let { stime, etime, time } = timeHandle(detailInfo.startTime, endTime, detailInfo.duration, timeRange);
            result.startTime = stime;
            result.endTime = etime;
            result.time = time;
            result.type = detailInfo.type;
            result.eventType = detailInfo.type + detailInfo.operate;
        } else {
            result.type = detailInfo.type;
            result.eventType = detailInfo.type;
        }
        detailInfo.startTime = formatTimsToMS(detailInfo.startTime);
        detailInfo.duration = formatTimsToMS(detailInfo.duration);
        result.info = detailInfo;
        return result;
    }
}

const BigType = {
    '0': 'on', 
    '1': 'file', 
    '2': 'net', 
    '3': 'futex', 
    '4': 'idle', 
    '5': 'other',
    '6': 'epoll'
}
export const dataHandle = (data: any, timeRange, trace: any) => {
    let {trace_id, src_ip, src_port, dst_ip, dst_port} = trace.labels;
    let requestStartTimestamp = Math.floor(trace.timestamp / 1000000);
    let totalTime: any = _.find(trace.metrics, {Name: 'request_total_time'});
    let requestEndTimestamp = Math.floor((trace.timestamp + totalTime.Data.Value) / 1000000);
    // let requestEndTimestamp = Math.floor((trace.timestamp + trace.metrics.request_total_time) / 1000000);
    let result: IThread[] = [];
    const groupData = _.groupBy(data, 'tid');
    const eventlist = _.cloneDeep(eventList);
    _.forEach(groupData, (list, key) => {
        let threadObj:IThread = {
            pid: list[0].pid,
            tid: list[0].tid,
            name: list[0].threadName,
            transactionId: list[0].transactionId,
            startTime: formatTimsToMS(list[0].startTime),
            endTime: formatTimsToMS(list[list.length - 1].endTime),
            eventList: [],
            javaLockList: [],
            logList: [],
            traceList: []
        };
        // cpuEvents 和 javaFutexEvents中存在不同数据段内返回startTime完全一致的数据，需要对重复数据进行过滤
        let cpuEvents = _.chain(list).map('cpuEvents').flatten().uniqBy('startTime').value();
        let javaFutexEvents = _.chain(list).map('javaFutexEvents').flatten().uniqBy('startTime').value();
        let transactionIdsList = _.chain(list).map('transactionIds').flatten().value();
        _.forEach(cpuEvents, event => {
            let { startTime } = event;
            let timeTypeList = _.compact(event.timeType.split(','));
            let timeValueList = _.compact(event.typeSpecs.split(',')).map((v: any) => parseFloat(v));
            // 可能出现对应事件0 没有log输出的情况，日志格式为log1||log3，所以不能用compact清除空值。 onInfo和offInfo同上
            let logList = event.log.split('|');
            let stackList = event.stack ? event.stack.split('|') : [];
            let onInfoList = event.onInfo.split('|');
            let offInfoList = event.offInfo.split('|');
            let onFlag = 0;
            let offFlag = 0;
            timeTypeList.forEach((type: any, idx) => {
                let endTime = startTime + timeValueList[idx];
                if (containTime(timeRange,startTime, endTime)) {
                    let { stime, etime, time } = timeHandle(startTime, endTime, timeValueList[idx], timeRange);
                    let eventObj: IEventTime = {
                        startTime: stime,
                        endTime: etime,
                        time: time,
                        type: BigType[type],
                        stackList: []
                    };
                    
                    if (type === '0') {
                        // TODO 日志其实需求根据@前面的数字截取字符串长度
                        if (logList.length > 0 && logList[onFlag]) {
                            let logInfo = logList[onFlag].split('@');
                            if (logInfo[1] && logInfo[1].length > 0) {
                                let traceId;
                                if (logInfo[1].length > 0) {
                                    let traceInfo = logInfo[1].match(/(?<=\[)(.+?)(?=\])/g);
                                    traceId = traceInfo ? _.trim(traceInfo[0].split(':')[1]) : '';
                                } else {
                                    traceId = '';
                                }
                                let logItem: ILogEvent = {
                                    eventType: 'log',
                                    startTime: formatTimsToMS(startTime),
                                    traceId: traceId,
                                    log: logInfo[1].length > 0 ? logInfo[1] : '--'
                                };
                                if (logItem.startTime > timeRange[0] && logItem.startTime < timeRange[1]) {
                                    threadObj.logList.push(logItem);
                                } 
                            } 
                            // else {
                            //     console.log(logInfo);
                            // }
                        }
                        // 火焰图数据处理
                        if (stackList.length > 1) {
                            let functionList = stackList[0].split('#');
                            // stackList 第一个|前面放的是function函数字段集合
                            if (stackList[onFlag + 1]) {
                                let stackInfo = stackList[onFlag + 1].split('#');
                                stackInfo.forEach(opt => {
                                    let temp = opt.split('-');
                                    let stackItem: any = {
                                        depth: parseInt(temp[0]),
                                        index: parseInt(temp[1]),
                                        width: parseInt(temp[2]),
                                        color: parseInt(temp[3]),
                                        name: functionList[temp[4]],
                                    };
                                    eventObj.stackList.push(stackItem);
                                });
                            } 
                        }
                        if (onInfoList.length > 0 && onInfoList[onFlag]) {
                            let result = onOffInfoHandle(onInfoList[onFlag], eventObj, timeRange);
                            if (_.isArray(result)) {
                                threadObj.eventList = [...threadObj.eventList, ...result];
                            } else {
                                if (result.endTime > timeRange[0] && result.startTime < timeRange[1]) {
                                    threadObj.eventList.push(result);
                                }
                            }
                        } else {
                            eventObj.eventType = eventObj.type;
                            threadObj.eventList.push(eventObj);
                        }
                        onFlag++;
                    }
                    if (type !== '0') {
                        if (offInfoList.length > 0 && offInfoList[offFlag]) {
                            let result = onOffInfoHandle(offInfoList[offFlag], eventObj, timeRange);
                            if (_.isArray(result)) {
                                threadObj.eventList = [...threadObj.eventList, ...result];
                            } else {
                                if (result.endTime > timeRange[0] && result.startTime < timeRange[1]) {
                                    threadObj.eventList.push(result);
                                }
                            }
                        } else {
                            eventObj.eventType = eventObj.type;
                            threadObj.eventList.push(eventObj);
                        }
                        offFlag++;
                    }
                    // 在时间筛选范围外的事件过滤掉不绘制
                    // if (eventObj.endTime > timeRange[0] && eventObj.startTime < timeRange[1]) {
                    //     threadObj.eventList.push(eventObj);
                    // }
                }
                startTime = endTime;
            });
        });
        // 判断trace中两个trace时间段内与当前trace的src_ip，src_port, dst_ip，dst_port相同的操作地址
        let operateFile = `${src_ip}:${src_port}->${dst_ip}:${dst_port}`;
        let netEventList = _.filter(threadObj.eventList, event => event.type === 'net');
        let activeNetTime: any = null;
        let activeNetType: any = null;
        if (netEventList.length > 0) {
            _.forEach(threadObj.eventList, event => {
                if (event.type === 'net') {
                    if (event.info && event.info.file === operateFile) {
                        console.log(operateFile, requestStartTimestamp, requestEndTimestamp);
                        // console.log('current event', event);
                        if (netReadTypes.indexOf(event.info.operate) > -1 && event.startTime > requestStartTimestamp - 2 && event.startTime < requestStartTimestamp + 2 ) {
                            // console.log('current event read', event);
                            event.active = true;
                            activeNetTime = event.startTime;
                            activeNetType = 1;
                        }
                        if (netWriteTypes.indexOf(event.info.operate) > -1 && event.startTime > requestEndTimestamp - 10 && event.startTime < requestEndTimestamp + 10 ) {
                            // console.log('current event write', event);
                            event.active = true;
                            activeNetTime = event.startTime;
                            activeNetType = 0;
                        }
                    }
                }
            });
        }

        if (activeNetTime && !trace_id) {
            _.forEach(transactionIdsList, item => {
                let tempTime = formatTimsToMS(item.timestamp);
                if (tempTime < activeNetTime + 1 && activeNetTime - 1 < tempTime && activeNetType === item.isEntry) {
                    trace_id = item.traceId;
                }
            });
        }

        // traceList处理
        if (transactionIdsList.length > 0) {
            const sameTraceList = _.uniqBy(_.filter(transactionIdsList, item => item.traceId === trace_id), 'timestamp');
            console.log('sameTraceList', sameTraceList);
            for(let i = 0;i< sameTraceList.length; i++) {
                if (i % 2 === 0 && sameTraceList[i+1]) {
                    threadObj.traceList.push({
                        traceId: sameTraceList[i].traceId,
                        startTime: formatTimsToMS(sameTraceList[i].timestamp),
                        endTime: formatTimsToMS(sameTraceList[i + 1].timestamp),
                        time: formatTimsToMS(sameTraceList[i + 1].timestamp - sameTraceList[i].timestamp),
                    });
                }
            }
            // console.log(threadObj.name, threadObj.traceList);
        }
        // java lock数据处理
        if (javaFutexEvents && javaFutexEvents.length > 0) {
            _.forEach(javaFutexEvents, lockEvent => {
                let { stime, etime, time } = timeHandle(lockEvent.startTime, lockEvent.endTime, lockEvent.endTime - lockEvent.startTime, timeRange);
                let infoList = lockEvent.dataValue.split('!');
                let waitTid = infoList[7] !== '-1' ? infoList[7] : '';
                let lastInfo = infoList[infoList.length - 1].replace(/\n/g, '');
                let javalock: IJavaLock = {
                    threadTid: threadObj.tid,
                    startTime: stime,
                    endTime: etime,
                    time: time,
                    type: 'lock',
                    lockAddress: infoList[3],
                    eventType: infoList[4],
                    stack: lastInfo ? lastInfo : infoList[infoList.length - 2],
                    waitThread: waitTid ? (groupData[waitTid] ? (groupData[waitTid] as any)[0].threadName : '--'): '--'
                };
                if (javalock.endTime > timeRange[0] && javalock.startTime < timeRange[1]) {
                    // javalock可能占据好几个事件时间段，原始数据存在startTime、endtime相同的lock事件，需要过滤掉
                    if (_.findIndex(threadObj.javaLockList, {startTime: javalock.startTime, endTime: javalock.endTime}) === -1) {
                        threadObj.javaLockList.push(javalock);
                    }
                }
            });
        }
        // 各事件类型数量统计
        _.forEach(eventlist, opt => {
            let num;
            if (opt.type === 'lock') {
                num = _.filter(threadObj.javaLockList, temp => temp.eventType === opt.value).length;
            } else {
                num = _.filter(threadObj.eventList, temp => temp.eventType === opt.value).length;
            }
            opt.count ? (opt.count += num) : (opt.count = num);
        });
        result.push(threadObj);
    });
    // 判断当前线程日志中解析的traceId是否包含trace数据中traceId
    _.forEach(result, item => {
        let logTraceIdList = _.chain(item.logList).map('traceId').compact().value();
        if (logTraceIdList.indexOf(trace_id) > -1) {
            item.active = true;
        }
    });

    // 统计对应线程中在request请求时间段内与trace关联的线程内的事件时间统计
    const allInfo: any[] = [];
    if (_.some(result, item => item.traceList.length > 0)) {
        let traceData: IThread[] = _.filter(result, item => item.traceList.length > 0) as IThread[];
        let requestEvent: IEventTime[] = [];
        let locks: IJavaLock[] = [];
        _.forEach(traceData, (item: IThread) => {
            _.forEach(item.eventList, event => {
                if (event.endTime > requestStartTimestamp - 2 && event.startTime < requestEndTimestamp + 2) {
                    let nEvent = _.cloneDeep(event);
                    let {stime, etime, time} = timeHandle(event.startTime, event.endTime, event.time, [requestStartTimestamp - 2, requestEndTimestamp + 2], false);
                    nEvent.startTime = stime;
                    nEvent.endTime = etime;
                    nEvent.time = time;
                    requestEvent.push(nEvent)
                }
            });
            _.forEach(item.javaLockList, lock => {
                if (lock.endTime > requestStartTimestamp && lock.startTime < requestEndTimestamp) {
                    let nLock = _.cloneDeep(lock);
                    let {stime, etime, time} = timeHandle(lock.startTime, lock.endTime, lock.time, [requestStartTimestamp, requestEndTimestamp], false);
                    nLock.startTime = stime;
                    nLock.endTime = etime;
                    nLock.time = time;
                    locks.push(nLock);
                }
            });
        });
        let requestEventByType = _.groupBy(requestEvent, 'type');
        _.forEach(requestEventByType, (list, key) => {
            let sevent: IEvent = _.find(eventList, {type: key}) as IEvent;
            let time = parseFloat(_.sum(_.map(list, 'time')).toFixed(2));
            const timeObj = {
                type: key,
                eventType: key,
                color: sevent.color,
                time,
                timeRate: ((time / (requestEndTimestamp - requestStartTimestamp)) * 100).toFixed(2)
            }
            if (timeObj.eventType === 'on') {
                allInfo.unshift(timeObj);
            } else {
                allInfo.push(timeObj);
            }
        });
        if (locks.length > 0) {
            let time = _.sum(_.map(locks, 'time'));
            let timeObj = {
                type: 'lock',
                eventType: 'lock',
                color: '#c0b81c',
                time: time.toFixed(2),
                timeRate: ((time / (requestEndTimestamp - requestStartTimestamp)) * 100).toFixed(2)
            }
            allInfo.push(timeObj);

            let waitObj = _.find(allInfo, {type: 'futex'});
            waitObj.time = parseFloat((parseFloat(waitObj.time) - time).toFixed(2));
            waitObj.timeRate = ((waitObj.time / (requestEndTimestamp - requestStartTimestamp)) * 100).toFixed(2)
        }
        console.log(allInfo);
    }
    
    // 获取整个trace过程中，trace处理的开始值跟结束值
    let allTraceList = _.chain(result).map('traceList').flatten().value();
    let traceTimes: number[] = [];
    if (allTraceList.length > 0) {
        let minStart = _.min(_.map(allTraceList, 'startTime'));
        let maxEnd = _.max(_.map(allTraceList, 'endTime'));
        traceTimes = [minStart, maxEnd];
    }
    // console.log(traceTimes);
    return {data: result, eventlist, traceTimes, requestInfo: allInfo, traceId: trace_id};
}

export const getLineTimesList = (requestTimes, traceTimes) => {
    const result: ILineTime[] = [];
    requestTimes.forEach(time => {
        result.push({
            time: new Date(time),
            type: 'request'
        });
    });
    traceTimes.forEach(time => {
        result.push({
            time: new Date(time),
            type: 'trace'
        });
    });
    return result;
}
// 构造线程筛选需要用的筛选数据
export const buildFilterList = (list: IThread[]) => {
    let threadList: any[] = [];
    let fileList: string[] = [];
    let fileOptList: any[] = [];
    list.forEach(item => {
        threadList.push({
            name: item.name,
            tid: item.tid
        });
        let files = _.chain(item.eventList).filter(opt => opt.type === 'file').map(opt => opt.info && opt.info.file).value();
        fileList = fileList.concat(files);
    });
    fileList = _.uniq(fileList);
    fileList.forEach(file => {
        fileOptList.push({
            label: file, 
            value: file
        })
    });
    // console.log(threadList, fileOptList);
    return {threadList, fileList: fileOptList};
}