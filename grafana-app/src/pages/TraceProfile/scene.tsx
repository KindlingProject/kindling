import React from 'react';
import TracePanel, { PanelOptions } from './panel';
import { PanelPlugin } from '@grafana/data';
import { sceneUtils, EmbeddedScene, SceneFlexLayout, SceneFlexItem, VizPanel } from '@grafana/scenes';

const profilePlugin = new PanelPlugin<PanelOptions>(TracePanel).useFieldConfig({
    useCustomConfig: (builder) => {
        builder.addBooleanSwitch({
            path: 'showLatency',
            name: 'Latency Config',
            description: 'The value change latency threshold config',
            defaultValue: false,
        });
    }
});
sceneUtils.registerRuntimePanelPlugin({ pluginId: 'profile-panel', plugin: profilePlugin });

function getScene() {
    return new EmbeddedScene({
        body: new SceneFlexLayout({
            children: [
                new SceneFlexItem({
                    width: '100%',
                    height: '100%',
                    body: new VizPanel({
                        pluginId: 'profile-panel',
                    })
                }),
            ]
        }),
    });
}

export const TracePluginPage = () => {
    const scene = getScene();

    return <scene.Component model={scene} />;
};
