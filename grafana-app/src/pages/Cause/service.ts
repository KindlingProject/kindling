import dayjs from "dayjs";
import _ from 'lodash';

export const flameDataHandle = (data: string) => {
    const stackList: any[] = [];
    const list = _.flatten(data.split('|'));
    const functionList = list[0].split('#');
    let stackInfo = list[1].split('#');
    stackInfo.forEach(opt => {
        let temp = opt.split('-');
        let stackItem: any = {
            depth: parseInt(temp[0], 10),
            index: parseInt(temp[1], 10),
            width: parseInt(temp[2], 10),
            color: parseInt(temp[3], 10),
            name: functionList[parseInt(temp[4], 10)],
        };
        stackList.push(stackItem);
    });
    return stackList;
}

export const basiclineOption: any = {
    title: {
        show: false,
        text: '',
        textStyle: {
            fontSize: 12
        },
    },
    tooltip: {
        trigger: 'axis',
        order: 'seriesAsc',
        confine: true,
    },
    legend: {
        show: true,
        icon: 'roundRect',
        type: 'scroll',
        left: 10,
        top: 0,
        itemWidth: 10,
        itemHeight: 10,
        data: []
    },
    grid: {
        left: 20,
        top: 40,
        right: 10,
        bottom: 20,
        containLabel: true,
    },
    xAxis: [{
        type: 'category',
        boundaryGap: false,
        data: [],
        axisLabel: {
            formatter: (v: any) => dayjs(parseInt(v, 10)).format('MM-DD HH:mm')
        }
    }],
    yAxis: [{
        type: 'value',
        axisLabel: {},
    }],
    series: []
};

export const buildLineSeriesData = (name: string, data: any, areaStyle: object = { opacity: 0.12 }, type = 'line') => {
    return {
        name: name,
        type: type,
        data: data,
        areaStyle: areaStyle,
        smooth: false,
        connectNulls: true,
        showSymbol: false,
    };
};
