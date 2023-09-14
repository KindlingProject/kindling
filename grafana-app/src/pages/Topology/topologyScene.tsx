import React, { useMemo } from 'react';
import TopologyPanel from './topologyPanel';
import { PrometheusName, ROUTES } from '../../constants';
import { TopologyOptions } from './types';
import { prefixRoute } from '../../utils/utils.routing';
import { PanelPlugin, VariableRefresh } from '@grafana/data';
import { getDataSourceSrv } from '@grafana/runtime';
import { QueryVariable, SceneVariableSet, VariableValueSelectors, SceneTimeRange, SceneControlsSpacer, SceneTimePicker, SceneRefreshPicker, SceneQueryRunner,
    sceneUtils, EmbeddedScene, SceneFlexLayout, SceneFlexItem, VizPanel, SceneApp, SceneAppPage } from '@grafana/scenes';

const Topology = new PanelPlugin<TopologyOptions>(TopologyPanel).useFieldConfig({
    useCustomConfig: (builder) => {
        builder.addBooleanSwitch({
            path: 'showLatency',
            name: 'Latency Config',
            description: 'The value change latency threshold config',
            defaultValue: true,
        })
        .addNumberInput({
            path: 'normalLatency',
            defaultValue: 20,
            name: 'Normal',
            showIf: config => config.showLatency,
        })
        .addNumberInput({
            path: 'abnormalLatency',
            defaultValue: 1000,
            name: 'Abnormal',
            showIf: config => config.showLatency,
        })
        .addBooleanSwitch({
            path: 'showRtt',
            name: 'RTT Config',
            description: 'The value change rtt threshold config',
            defaultValue: false,
        })
        .addNumberInput({
            path: 'normalRtt',
            defaultValue: 100,
            name: 'Normal',
            showIf: config => config.showRtt,
        })
        .addNumberInput({
            path: 'abnormalRtt',
            defaultValue: 200,
            name: 'Abnormal',
            showIf: config => config.showRtt,
        })
    }
});
sceneUtils.registerRuntimePanelPlugin({ pluginId: 'topology-panel', plugin: Topology });

const prometheusDataSource = getDataSourceSrv().getInstanceSettings(PrometheusName);
const prometheusUid = prometheusDataSource?.uid;

const namespace = new QueryVariable({
    name: 'Namespace',
    datasource: {
        type: 'prometheus',
        uid: prometheusUid,
    },
    query: {
        query: 'kindling_topology_request_total'
    },
    regex: '.*namespace=\"(.*?)\".*',
    refresh: VariableRefresh.onTimeRangeChanged
});

const workload = new QueryVariable({
    name: 'Workload',
    datasource: {
        type: 'prometheus',
        uid: prometheusUid,
    },
    query: {
        query: 'label_values(kindling_k8s_workload_info{namespace="$Namespace"}, workload_name)'
    }
});

const queryRunner = new SceneQueryRunner({
    datasource: {
        type: 'prometheus',
        uid: prometheusUid,
    },
    queries: [
        {
            refId: 'A',
            expr: 'increase(kindling_topology_request_total{src_namespace=~"$Namespace", src_workload_name=~"$Workload"}[$__range]) or increase(kindling_topology_request_total{dst_namespace=~"$Namespace", dst_workload_name=~"$Workload"}[$__range])',
            instant: true
        },
        {
            refId: 'B',
            expr: 'increase(kindling_topology_request_duration_nanoseconds_total{src_namespace=~"$Namespace", src_workload_name=~"$Workload"}[$__range]) or increase(kindling_topology_request_duration_nanoseconds_total{dst_namespace=~"$Namespace", dst_workload_name=~"$Workload"}[$__range])',
            instant: true
        },
        {
            refId: 'C',
            expr: 'sum (increase(kindling_entity_request_total{namespace=~"$Namespace", workload_name=~"$Workload"}[$__range])) by (namespace, workload_name, pod)',
            instant: true
        },
        {
            refId: 'D',
            expr: 'sum (increase(kindling_entity_request_duration_nanoseconds_total{namespace=~"$Namespace", workload_name=~"$Workload"}[$__range])) by (namespace, workload_name, pod)',
            instant: true
        },
        {
            refId: 'E',
            expr: 'sum(increase(kindling_entity_request_send_bytes_total{namespace=~"$Namespace", workload_name=~"$Workload"}[$__range])) by(namespace, workload_name, pod)',
            instant: true
        },
        {
            refId: 'F',
            expr: 'sum(increase(kindling_entity_request_receive_bytes_total{namespace=~"$Namespace", workload_name=~"$Workload"}[$__range])) by(namespace, workload_name, pod)',
            instant: true
        },
    ],
});

function getScene() {
    const timeRange = new SceneTimeRange({
        from: 'now-6h',
        to: 'now',
    });

    return new EmbeddedScene({
        $timeRange: timeRange,
        $data: queryRunner,
        $variables: new SceneVariableSet({
            variables: [namespace, workload],
        }),
        body: new SceneFlexLayout({
            children: [
                new SceneFlexItem({
                    width: '100%',
                    height: '100%',
                    body: new VizPanel({
                        pluginId: 'topology-panel',
                        options: {
                            showLatency: true,
                            abnormalLatency: 1000,
                            abnormalRtt: 200,
                            normalLatency: 20,
                            normalRtt: 100,
                        }
                    })
                }),
            ]
        }),
        controls: [
            new VariableValueSelectors({}),
            new SceneControlsSpacer(),
            new SceneTimePicker({ isOnCanvas: true }),
            new SceneRefreshPicker({ isOnCanvas: true }),
        ],
    });
}

function getSceneAppPage() {
    return new SceneApp({
        pages: [
            new SceneAppPage({
                title: 'Kindling Topology',
                url: prefixRoute(ROUTES.Topology),
                getScene: () => {
                    return getScene();
                },
            }),
        ],
    });
}

export const TopologyPluginPage = () => {
    // const scene = getScene();
    const scene = useMemo(() => getSceneAppPage(), []);

    return <scene.Component model={scene} />;
};
