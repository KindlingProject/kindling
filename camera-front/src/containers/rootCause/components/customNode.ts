import G6 from '@antv/g6';
import { formatUnit } from '@/services/util';

const nodeW = 180;
// 注册自定义节点
G6.registerNode('call-node', {
    draw: (node: any, group: any) => {
        
        let shape = group.addShape('rect', {
            attrs: {
                x: 0,
                y: 0,
                width: 200,
                height: node.list.length > 2 ? node.list.length * 15 + 25 : 55,
                fill: '#ffffff',
                stroke: node.is_warn ? '#ff3c3c' : '#9d9d9d',
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
                fontSize: 8,
                lineHeight: 10,
                text: `${node.id}`,
                fill: '#595959',
            },
            id: 'node-name',
            name: 'node-name',
        });
        // group.addShape('text', {
        //     attrs: {
        //         x: 5,
        //         y: 15,
        //         textBaseline: 'top',
        //         class: 'node_text',
        //         fontSize: 8,
        //         lineHeight: 10,
        //         text: `url: ${node.content_key}`,
        //         fill: '#595959',
        //     },
        //     id: 'node-name',
        //     name: 'node-name',
        // });

        // let nameTextBox = nameText.getBBox();
        // shape.attr({
        //     width: nameTextBox.width + 10
        // });


        node.list.forEach((opt, idx) => {
            group.addShape('text', {
                attrs: {
                    x: nodeW / 4,
                    y: 15 + (idx + 1) * 12,
                    textAlign: 'center',
                    textBaseline: 'top',
                    class: 'node_text',
                    fontSize: 8,
                    lineHeight: 10,
                    text: `历史调用P90: ${formatUnit(opt.p90, 'ns')}`,
                    timeIdx: idx,
                    cursor: opt.is_profiled ? 'pointer' : 'default',
                    fill: opt.is_profiled ? '#096dd9' : '#595959',
                },
                name: 'node-time-text',
            });
            group.addShape('text', {
                attrs: {
                    x: nodeW / 4 * 3,
                    y: 15 + (idx + 1) * 12,
                    textAlign: 'center',
                    textBaseline: 'top',
                    class: 'node_text',
                    fontSize: 8,
                    lineHeight: 10,
                    text: `本次故障调用：${formatUnit(opt.totalTime, 'ns')}`,
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
