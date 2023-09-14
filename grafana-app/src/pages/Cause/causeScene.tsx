import React, { useMemo } from 'react';
import CustomVizPanel, { CustomVizOptions } from './causeVizPanel';
import TracePanel from './TracePanel';
import { PanelPlugin } from '@grafana/data';
import { ROUTES } from '../../constants';
import { prefixRoute } from '../../utils/utils.routing';
import { sceneUtils, EmbeddedScene, SceneFlexLayout, SceneFlexItem, VizPanel, SceneApp, SceneAppPage, 
    TextBoxVariable, SceneVariableSet, SceneTimeRange, VariableValueSelectors, SceneTimePicker, SceneControlsSpacer} from '@grafana/scenes';

const CausePanel = new PanelPlugin<CustomVizOptions>(CustomVizPanel).useFieldConfig({
    useCustomConfig: (builder) => {
        builder.addNumberInput({
            path: 'testOption',
            name: 'Test option',
            defaultValue: 1,
        });
    },
});
sceneUtils.registerRuntimePanelPlugin({ pluginId: 'cause-panel', plugin: CausePanel });

function getScene(traceId: string) {
    return new EmbeddedScene({
        body: new SceneFlexLayout({
            children: [
                new SceneFlexItem({
                    width: '100%',
                    height: '100%',
                    body: new VizPanel({
                        title: '故障根因推导',
                        pluginId: 'cause-panel',
                        options: {
                            traceId: traceId
                        }
                    })
                }),
            ]
        }),
    });
}

const TracePluginPanel = new PanelPlugin(TracePanel);
sceneUtils.registerRuntimePanelPlugin({ pluginId: 'trace-panel', plugin: TracePluginPanel });

function getTraceScene() {
    const pidVariable = new TextBoxVariable({
        name: 'pid',
        label: 'Pid',
        value: '',
    });
    const urlVariable = new TextBoxVariable({
        name: 'url',
        label: 'URL',
        value: '',
    });
    const traceIdVariable = new TextBoxVariable({
        name: 'traceId',
        label: 'TraceId',
        value: '',
    });
    const timeRange = new SceneTimeRange({
        from: 'now-30m',
        to: 'now',
    });

    return new EmbeddedScene({
        $timeRange: timeRange,
        $variables: new SceneVariableSet({ variables: [pidVariable, urlVariable, traceIdVariable] }),
        body: new SceneFlexLayout({
            children: [
                new SceneFlexItem({
                    width: '100%',
                    height: '100%',
                    body: new VizPanel({
                        title: 'Trace List',
                        pluginId: 'trace-panel'
                    })
                }),
            ]
        }),
        controls: [
            new VariableValueSelectors({}),
            new SceneControlsSpacer(),
            new SceneTimePicker({ isOnCanvas: true })
        ],
    });
}
function getSceneAppPage() {
    return new SceneApp({
        pages: [
            new SceneAppPage({
                title: '',
                url: prefixRoute(ROUTES.Cause),
                getScene: () => {
                    return getTraceScene();
                },
                drilldowns: [
                    {
                        routePath: prefixRoute(`${ROUTES.Cause}`) + '/report/:traceId',
                        getPage(routeMatch, parent) {
                            const traceId = routeMatch.params.traceId;

                            return new SceneAppPage({
                                url: prefixRoute(`${ROUTES.Cause}`) + `/report/${traceId}`,
                                title: '',
                                getParentPage: () => parent,
                                getScene: () => {
                                    return getScene(traceId);
                                },
                            });
                        },
                    }
                ]
            }),
        ]
    });
}

export const CausePluginPage = () => {
    // const scene = getScene();
    const scene = useMemo(() => getSceneAppPage(), []);

    return <scene.Component model={scene} />;
};
