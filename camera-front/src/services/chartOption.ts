import moment from "moment";

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
            formatter: (v) => moment(parseInt(v)).format('MM-DD HH:mm')
        }
    }],
    yAxis: [{
        type: 'value',
        axisLabel: {},
    }],
    series: []
};

export const lineWithoutAxis: any = {
    tooltip: {
        trigger: 'axis',
    },
    grid: {
        left: 0,
        top: 1,
        right: 0,
        bottom: 1,
    },
    xAxis: [{
        show: false,
        type: 'category',
    }],
    yAxis: [{
        show: false
    }],
    series: []
}

export const buildLineSeriesData = (name, data, areaStyle: object = { opacity: 0.12 }, type = 'line') => {
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
