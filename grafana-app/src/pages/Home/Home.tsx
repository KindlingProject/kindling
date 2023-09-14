import React, { useMemo } from 'react';

import { SceneApp, SceneAppPage } from '@grafana/scenes';
import { getBasicScene } from './scenes';
import { prefixRoute } from '../../utils/utils.routing';
import { DATASOURCE_REF, ROUTES } from '../../constants';
import { config } from '@grafana/runtime';
import { Alert } from '@grafana/ui';

const getScene = () => {
  return new SceneApp({
    pages: [
      new SceneAppPage({
        title: 'Home page',
        subTitle:
          'This scene showcases a basic scene functionality, including query runner, variable and a custom scene object.',
        url: prefixRoute(ROUTES.Home),
        getScene: () => {
          return getBasicScene();
        },
      }),
    ],
  });
};
export const HomePage = () => {
  const scene = useMemo(() => getScene(), []);

  return (
    <>
      {!config.featureToggles.topnav && (
        <Alert title="Missing topnav feature toggle">
          Scenes are designed to work with the new navigation wrapper that will be standard in Grafana 10
        </Alert>
      )}

      {!config.datasources[DATASOURCE_REF.uid] && (
        <Alert title={`Missing ${DATASOURCE_REF.uid} datasource`}>
          These demos depend on <b>testdata</b> datasource: <code>{JSON.stringify(DATASOURCE_REF)}</code>. See{' '}
          <a href="https://github.com/grafana/grafana/tree/main/devenv#set-up-your-development-environment">
            https://github.com/grafana/grafana/tree/main/devenv#set-up-your-development-environment
          </a>{' '}
          for more details.
        </Alert>
      )}

      <scene.Component model={scene} />
    </>
  );
};
