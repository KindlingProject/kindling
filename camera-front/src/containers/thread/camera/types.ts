// 绘制图表时配置option 
export interface IOption {
    data: IThread[];
    trace: any;
    traceId: string;
    requestTime?: Date[];
    lineTimeList: any[];
    timeRange: Date[];
    svgId?: string;
    nameWidth?: number;
    barWidth?: number;
    barPadding?: number;
    padding?: number;
    parentRef?: any;
    eventClick: (evt: any) => void;
    nameClick: (tid: string) => void;
}

export type IEventName = 'ON' | 'DISK' | 'NET' | 'FUTEX' | 'IDLE' | 'OHTER';
export type IEventType = '0' | '1' | '2' | '3' | '4' | '5';

export interface IFilterParams {
    threadList: number[];
    logList: string[];
    fileList: string[];
    eventList: string[];
}

export type IEvent = {
    type: string;
    name: string;
    value: string;
    color: string;
    fillColor: string;
    activeColor: string;
    count?: number;
}

export type IEventTime  = {
    startTime: number;
    endTime: number;
    time: number;
    type: string;
    eventType?: string;
    name?: string;
    info?: any;
    stackList: any[];
    active?: boolean;
    message?: any;
    // 后续处理数据判断虚线是否需要合并
    idx?: number;
    timeRate?: number;
    left?: number;
    children?: IEventTime[];
}
export type IJavaLock  = {
    threadTid: number;
    startTime: number;
    endTime: number;
    time: number;
    type: string;
    eventType: string;
    lockAddress: string;
    stack: string;
    waitThread: string;
}
export type ILogEvent  = {
    eventType: string;
    startTime: number;
    traceId: string;
    log: string;
}
export interface IThread {
    pid: number;
    tid: number;
    name: string;
    active?: boolean;
    transactionId?: string;
    startTime: number;
    endTime: number;
    traceStartTimestamp?: number;
    traceEndTimestamp?: number;
    eventList: IEventTime[];
    javaLockList: IJavaLock[];
    logList: ILogEvent[];
    traceList: any[]
}

export type ILineTime = {
    time: Date,
    type: 'request' | 'trace'
}

export type ILegend = {
    name: string;
    color: string;
}