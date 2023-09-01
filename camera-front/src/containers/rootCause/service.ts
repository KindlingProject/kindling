import _ from 'lodash';

export const flameDataHandle = (data: string) => {
    const stackList: any[] = [];
    const list = _.flatten(data.split('|'));
    const functionList = list[0].split('#');
    let stackInfo = list[1].split('#');
    stackInfo.forEach(opt => {
        let temp = opt.split('-');
        let stackItem: any = {
            depth: parseInt(temp[0]),
            index: parseInt(temp[1]),
            width: parseInt(temp[2]),
            color: parseInt(temp[3]),
            name: functionList[temp[4]],
        };
        stackList.push(stackItem);
    });
    return stackList;
}