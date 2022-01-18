import * as G6 from '@antv/g6';

// custom the edge with an extra rect
G6.registerEdge('service-edge', {
    afterDraw(cfg: any, group: any) {
        // get the first shape in the graphics group of this edge, it is the path of the edge here
        // 获取图形组中的第一个图形，在这里就是边的路径图形
        const shape = group.get('children')[0];
        // get the coordinate of the mid point on the path
        // 获取路径图形的中点坐标
        const midPoint = shape.getPoint(0.5);
        const rectColor = cfg.midPointColor || '#FFF';
        // add a rect on the mid point of the path. note that the origin of a rect shape is on its lefttop
        // 在中点增加一个矩形，注意矩形的原点在其左上角
        group.addShape('rect', {
            attrs: {
                width: 120,
                height: 20,
                fill: rectColor,
                // x and y should be minus width / 2 and height / 2 respectively to translate the center of the rect to the midPoint
                // x 和 y 分别减去 width / 2 与 height / 2，使矩形中心在 midPoint 上
                x: midPoint.x - 5,
                y: midPoint.y - 5,
            },
        });
  
        // // get the coordinate of the quatile on the path
        // // 获取路径上的四分位点坐标
        // const quatile = shape.getPoint(0.25);
        // const quatileColor = cfg.quatileColor || '#FFF';
        // // add a circle on the quatile of the path
        // // 在四分位点上放置一个圆形
        // group.addShape('circle', {
        //     attrs: {
        //         r: 5,
        //         fill: quatileColor || '#333',
        //         x: quatile.x,
        //         y: quatile.y,
        //     },
        // });
    },
    update: undefined,
}, 'cubic');
