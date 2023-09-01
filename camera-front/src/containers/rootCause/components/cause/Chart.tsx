import React, { forwardRef, memo, useImperativeHandle, useRef } from 'react';
import { Empty } from 'antd';
import ReactEcharts from 'echarts-for-react';
import * as echarts from 'echarts';
import { isEqual } from 'lodash';

const emptyStyle: React.CSSProperties = {
    margin: 0,
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    flexDirection: 'column',
};

// echarts.registerTheme('myDark', myDark);
// echarts.registerTheme('myMacarons', myMacarons);

interface IProps {
    option: any;
    height?: string | number;
    events?: any;
    noDataImg?: any;
    className?: string;
}
function Chart(props: IProps, ref) {

    const {
        option,
        height = '100%',
        events = null,
        noDataImg = null,
        className = ''
    } = props;
    const echartRef: any = useRef(null);

    // 将echarts实例抛出
    useImperativeHandle(ref, () => {
        return {
            echartInstance: echartRef?.current?.getEchartsInstance() || null,
        }
    })

    return (
        option.series?.length > 0 && option.series[0]?.data?.length > 0 ? (
            <ReactEcharts
                ref={echartRef}
                className={className}
                option={option}
                style={{ height, width: '100%' }}
                onEvents={events}
                notMerge></ReactEcharts>
        ) : (
            <Empty
                image={noDataImg || Empty.PRESENTED_IMAGE_SIMPLE}
                style={{ height, ...emptyStyle }}
                description='暂无数据'></Empty>
        )
    );
}

const isOptEqual = (prevProps: IProps, nextProps: IProps) => {
    if (isEqual(prevProps.option, nextProps.option)) {
        return true;
    } else {
        return false;
    }
}
export default memo(forwardRef(Chart), isOptEqual);