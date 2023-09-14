import dayjs from 'dayjs';

export const dateTime = 'YYYY-MM-DD HH:mm:ss';
/**
 * [时间单位转换-毫秒]
 * 默认 ms
 * @param   {[type]}  time [time value]
 * @return  {[type]}       [time value with unit]
 */
export const formatTime = (time: any, reserve = 2) => {
    let flag = false;
    if (time === 'NaN' || time === undefined) {
        return 0;
    } else {
        flag = time < 0;
        time = Math.abs(time);
        if (time > 3600000) {
            time = (time / 3600000).toFixed(reserve) + 'h';
        } else if (time >= 60000) {
            time = (time / 60000).toFixed(reserve) + 'min';
        } else if (time >= 1000) {
            time = (time / 1000).toFixed(reserve) + 's';
        } else if (time >= 10) {
            time = time.toFixed(reserve) + 'ms';
        } else if (time > 0) {
            time = time.toFixed(reserve) + 'ms';
        } else {
            time = time.toFixed(reserve);
        }
        return flag ? '-' + time : time;
    }
};

export const formatCount = (y: number, reserve = 2) => {
    const WAN = 10000;
    const YI = 100000000;
    const ZHAO = 1000000000000;
    const yy: number = Math.abs(y);
    if (yy < WAN) {
        return yy;
    }
    if (yy < YI) {
        return `${(yy / WAN).toFixed(reserve)}${decodeURIComponent('%E4%B8%87')}`; // 万
    }
    if (yy < ZHAO) {
        return `${(yy / YI).toFixed(reserve)}${decodeURIComponent('%E4%BA%BF')}`; // 亿
    }
    return `${(yy / ZHAO).toFixed(reserve)}${decodeURIComponent('%E5%85%86')}`; // 兆
};

/**
 * 转化字节数到对应的量级(TB,GB,MB,KB)
 * @param {number} y 默认是B
 */
export const formatKMBT = (y: number, reserve = 2) => {
    const yy = Math.abs(y);
    if (yy >= Math.pow(1024, 4)) {
        return y < 0
            ? -1 * +(yy / Math.pow(1024, 4)).toFixed(reserve) + 'T'
            : (yy / Math.pow(1024, 4)).toFixed(reserve) + 'T';
    } else if (yy >= Math.pow(1024, 3)) {
        return y < 0
            ? -1 * +(yy / Math.pow(1024, 3)).toFixed(reserve) + 'G'
            : (yy / Math.pow(1024, 3)).toFixed(reserve) + 'G';
    } else if (yy >= Math.pow(1024, 2)) {
        return y < 0
            ? -1 * +(yy / Math.pow(1024, 2)).toFixed(reserve) + 'M'
            : (yy / Math.pow(1024, 2)).toFixed(reserve) + 'M';
    } else if (yy >= 1024) {
        return y < 0 ? -1 * +(yy / 1024).toFixed(reserve) + 'K' : (yy / 1024).toFixed(reserve) + 'K';
    } else if (yy < 1024 && yy >= 1) {
        return y < 0 ? -1 * +yy.toFixed(reserve) + 'B' : yy.toFixed(reserve) + 'B';
    } else if (yy < 1 && yy > 0) {
        return y < 0 ? -1 * +yy.toFixed(reserve) + 'B' : yy.toFixed(reserve) + 'B';
    } else if (yy === 0) {
        return 0;
    } else {
        return yy;
    }
};

/**
 * 转化字节数到对应的量级(TB,GB,MB)
 * @param {number} y 默认是MB
 */
export const formatMGT = (y: number) => {
    const yy = Math.abs(y);
    if (yy >= 1024 * 1024) {
        return y < 0
            ? -1 * +(yy / (1024 * 1024)).toFixed(2) + 'T'
            : (yy / (1024 * 1024)).toFixed(2) + 'T';
    } else if (yy >= 1024) {
        return y < 0 ? -1 * +(yy / 1024).toFixed(2) + 'G' : (yy / 1024).toFixed(2) + 'G';
    } else if (yy < 1024 && yy >= 1) {
        return y < 0 ? -1 * +yy.toFixed(2) + 'M' : yy.toFixed(2) + 'M';
    } else if (yy < 1 && yy > 0) {
        return y < 0 ? -1 * yy + 'M' : yy + 'M';
    } else if (yy === 0) {
        return 0;
    } else {
        return yy + 'M';
    }
};

//小数转化百分比
export const formatPercent = (number: number, reserve = 2) => {
    if (number !== 0 && !number) {
        return '-';
    }
    if (number === 0) {
        return 0;
    } else {
        if (number >= 0.01) {
            return `${parseFloat(number + '').toFixed(reserve)}%`;
        } else {
            return '<0.01%';
        }
    }
};

export type IUnit = 'byteMin' | 'byteMins' | 'byte' | 'KB' | 'KB/S' | 'ms' | 'us' | 'ns' | 'count' | '%' | '100%' | '核' | '2' | 'date' | '个' | '';
export const formatUnit = (data: any, unit: IUnit, reserve = 2) => {
    if (data === -1 || data === 'NaN') {
        return 'N/A';
    }
    if (data === undefined || data === null) {
        return '--';
    }
    if (data === '0' || data === 0) {
        return 0;
    }
    switch (unit) {
        case 'byteMin': //后台返回B
            return formatKMBT(data, reserve);
        case 'byteMins': //后台返回B
            return formatKMBT(data) ? formatKMBT(data).toString() + '/s' : 0;
        case 'byte': //后台返回单位MB
            return formatMGT(data);
        case 'KB': //后台返回单位KB
            return formatKMBT(data * 1024);
        case 'KB/S':
            return formatKMBT(data * 1024).toString() + '/S';
        case 'ms': //后台返回ms
            return formatTime(data, reserve);
        case 'us': //后台返回us
            return formatTime(data / 1000, reserve);
        case 'ns': //后台返回ns
            return formatTime(data / 1000000, reserve);
        case 'date': 
            return dayjs(data).format(dateTime);
        case 'count': //后台返回数量
            return formatCount(data, reserve);
        case '%': //百分数加单位
            return formatPercent(data, reserve);
        case '100%': //小数转换百分数
            return formatPercent(data * 100, reserve);
        case '核': {
            const v = parseFloat(data);
            if (v > 0 && v < 0.01) {
                if (v * 1000 < 0.01) {
                    return '<0.01微核';
                } else {
                    return (v * 1000).toFixed(2) + '微核';
                }
            }
            return `${v.toFixed(2)}核`;
        }
        case '2':
            return `${parseFloat(data).toFixed(2)}`;
        case '个':
            return `${data}个`;
        default:
            return `${data}`;
    }
};
