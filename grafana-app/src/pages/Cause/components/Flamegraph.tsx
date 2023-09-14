import React, { useEffect, useRef, useState } from 'react';
import { EmptySearchResult } from '@grafana/ui'; 
import * as d3 from 'd3';
import { flamegraph } from 'd3-flame-graph';
import 'd3-flame-graph/dist/d3-flamegraph.css';

interface TreeData {
  children?: TreeData[],
  depth: number;
  index: number;
  name: string;
  width: number;
  color: number;
  value: number;
  nextNodeIndex: number;
}

export const sleep = (n = 500) => {
  return new Promise((resolve) => setTimeout(resolve, n));
};


const sortData = (a: TreeData, b: TreeData) => {
  if (a.depth < b.depth) {
    return -1;
  } else if (a.depth > b.depth) {
    return 1;
  } else {
    return a.index - b.index;
  }
};
function transferTree(stackList: TreeData[]) {
  let treeData: any = {};
  if (!Array.isArray(stackList)) {
    return treeData;
  }
  stackList.sort(sortData);
  let tmp: TreeData;
  let nodeList: TreeData[] = [];
  stackList.forEach((item, index) => {
    if (item.depth === 0) {
      const obj: TreeData = {
        ...item,
        value: item.width,
        children: [],
        nextNodeIndex: index + 1
      };
      treeData = obj;
      nodeList.push(obj);
      tmp = treeData;
    } else {
      let flag = true;
      do {
        if ((item.depth === tmp.depth + 1) && item.index >= tmp.index && item.index < tmp.index + tmp.width) {
          flag = false;
          const obj: TreeData = {
            ...item,
            value: item.width,
            children: [],
            nextNodeIndex: index + 1
          };
          tmp.children?.push(obj);
          nodeList.push(obj);
        } else {
          tmp = nodeList[tmp.nextNodeIndex]
        }
      } while (flag && tmp);
    }
  })
  return treeData;
}
export default function Flamegraph(props: any) {

    const { data } = props;

    const ref = useRef<HTMLDivElement>(null);
    const [hoverCtx, setHoverCtx] = useState<any>(null);
    const init = useRef(true);
    let chart: any, renderFlag;

    const drawFlamegraph = (container: any, drawData: any) => {
        const width = container.offsetWidth - 10;
        d3.select('.flamegraph').selectAll('svg').remove();
        chart = flamegraph()
            .width(width)
            .cellHeight(18)
            .transitionDuration(750)
            .minFrameSize(5)
            // .transitionEase('0.5')
            // .sort(true)
            // .setColorHue('pastelgreen')
            // .onClick(onClick)
            //@ts-ignore  
            // 先注释掉不显示hover
            // .onHover(renderTooltip)
            .inverted(true)
            .selfValue(false);

        d3.select(`.flamegraph`).datum(drawData).call(chart)
        d3.select('.flamegraph').selectAll('g').on('mouseleave', async () => {
            renderFlag = false;
            await sleep(100);
            if (!renderFlag) {
                setHoverCtx(null);
            }
        });
    }

    // const onResize = () => {
    //     const dom = ref.current;
    //     if (dom) {
    //         drawFlamegraph(dom, data);
    //     }
    // }

    useEffect(() => {
        // window.addEventListener("resize", onResize);
        // console.log(data);
        if (data) {
            if (init.current) {
                init.current = false;
                const dom = ref.current;
                if (dom) {
                    drawFlamegraph(dom, transferTree(data));
                }
            } else if (chart) {
                chart.update(transferTree(data))
            }
        }
        return () => {
            d3.select('.flamegraph').selectAll('svg').remove();
        }
    // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [data])

    return (
        <div className="event_detail f-padding10">
            {
                data?.length ? <React.Fragment>
                    <div className="flamegraph" ref={ref} />
                    {hoverCtx}
                </React.Fragment> : <EmptySearchResult>Could not find anything matching your query</EmptySearchResult>
            }
        </div>
    );
}
