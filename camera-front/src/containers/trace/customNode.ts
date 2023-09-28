import G6 from '@antv/g6';
import { formatTimeNs } from '../thread/camera/util';

const nodeW = 180;
// 注册自定义节点
G6.registerNode('custom-node', {
    draw: (node: any, group: any) => {
        
        let shape = group.addShape('rect', {
            attrs: {
                x: 0,
                y: 0,
                width: nodeW,
                height: node.list.length > 2 ? node.list.length * 15 + 25 : 55,
                fill: '#ffffff',
                stroke: '#9d9d9d',
                radius: 4,
                fillOpacity: 0.08
            },
            id: 'node-rect',
            name: 'node-rect',
            draggable: true,
        });

        let nameText = group.addShape('text', {
            attrs: {
                x: 5,
                y: 5,
                textBaseline: 'top',
                class: 'node_text',
                fontSize: 10,
                lineHeight: 10,
                text: node.dst_pod,
                fill: '#595959',
            },
            id: 'node-name',
            name: 'node-name',
        });
        group.addShape('text', {
            attrs: {
                x: nodeW / 2,
                y: 17,
                textAlign: 'center',
                textBaseline: 'top',
                class: 'node_text',
                fontSize: 10,
                lineHeight: 10,
                text: node.pid,
                fill: '#595959',
            },
            id: 'node-name',
            name: 'node-name',
        });

        let nameTextBox = nameText.getBBox();
        shape.attr({
            width: nameTextBox.width + 10
        });


        node.list.forEach((opt, idx) => {
            group.addShape('text', {
                attrs: {
                    x: nodeW / 2,
                    y: 15 + (idx + 1) * 12,
                    textAlign: 'center',
                    textBaseline: 'top',
                    class: 'node_text',
                    fontSize: 10,
                    lineHeight: 10,
                    text: `${formatTimeNs(opt.totalTime)}   ${formatTimeNs(opt.p90)}`,
                    timeIdx: idx,
                    cursor: opt.is_profiled ? 'pointer' : 'default',
                    fill: opt.is_profiled ? '#096dd9' : '#595959',
                },
                name: 'node-time-text',
            });
        });
        
        return shape;
    },
    afterDraw(cfg: any, group) {
        // if (cfg.is_profiled) {
        //     let nodeTimeText = group?.findAllByName('node-time-text');
        //     nodeTimeText?.forEach(text => {
        //         text.on('mouseenter', () => {
        //             text.attr({ fill: '#096dd9' });
        //         });
        //         text.on('mouseleave', () => {
        //             text.attr({ fill: '#595959' });
        //         });
        //     });
        // }
    },
    update: (cfg: any, node: any) => {

    },
    afterUpdate(cfg, item) {

    },
    setState: (name, value, item) => {
      
    }
}, 'single-node');
