import React, { useMemo } from 'react';
import { prefixRoute } from '../../utils/utils.routing';
import { DATASOURCE_REF, ROUTES } from '../../constants';
import {
  EmbeddedScene,
  SceneApp,
  SceneAppPage,
  SceneFlexItem,
  SceneFlexLayout,
  SceneQueryRunner,
  SceneTimePicker,
  SceneTimeRange,
} from '@grafana/scenes';
import { getHumidityOverviewScene, getTemperatureOverviewScene } from './scenes';
import { getRoomsTemperatureStats, getRoomsTemperatureTable } from './panels';

const roomsTemperatureQuery = {
  refId: 'Rooms temperature',
  datasource: DATASOURCE_REF,
  scenarioId: 'random_walk',
  seriesCount: 8,
  alias: '__house_locations',
  min: 10,
  max: 27,
};

const getScene = () =>
  new EmbeddedScene({
    $data: new SceneQueryRunner({
      datasource: DATASOURCE_REF,
      queries: [roomsTemperatureQuery],
      maxDataPoints: 100,
    }),

    body: new SceneFlexLayout({
      direction: 'column',
      children: [
        new SceneFlexItem({
          height: 300,
          body: getRoomsTemperatureTable(),
        }),
        new SceneFlexItem({
          ySizing: 'fill',
          body: getRoomsTemperatureStats(),
        }),
      ],
    }),
  });

const getDrilldownsAppScene = () => {
  return new SceneApp({
    pages: [
      new SceneAppPage({
        $timeRange: new SceneTimeRange({ from: 'now-6h', to: 'now' }),
        title: 'Page with drilldown',
        subTitle: 'This scene showcases a basic drilldown functionality. Interact with room to see room details scene.',
        controls: [new SceneTimePicker({ isOnCanvas: true })],
        url: prefixRoute(`${ROUTES.WithDrilldown}`),
        hideFromBreadcrumbs: true,
        getScene,
        drilldowns: [
          {
            routePath: prefixRoute(`${ROUTES.WithDrilldown}`) + '/room/:roomName',
            getPage(routeMatch, parent) {
              const roomName = routeMatch.params.roomName;

              return new SceneAppPage({
                url: prefixRoute(`${ROUTES.WithDrilldown}`) + `/room/${roomName}/temperature`,
                title: `${roomName} overview`,
                subTitle: 'This scene is a particular room drilldown. It implements two tabs to organise the data.',
                getParentPage: () => parent,
                getScene: () => {
                  return new EmbeddedScene({ body: new SceneFlexLayout({ children: [] }) });
                },
                tabs: [
                  new SceneAppPage({
                    title: 'Temperature',
                    url: prefixRoute(`${ROUTES.WithDrilldown}`) + `/room/${roomName}/temperature`,
                    getScene: () => getTemperatureOverviewScene(roomName),
                  }),
                  new SceneAppPage({
                    title: 'Humidity',
                    url: prefixRoute(`${ROUTES.WithDrilldown}`) + `/room/${roomName}/humidity`,
                    getScene: () => getHumidityOverviewScene(roomName),
                  }),
                ],
              });
            },
          },
        ],
      }),
    ],
  });
};

export const WithDrilldown = () => {
  const scene = useMemo(() => getDrilldownsAppScene(), []);

  return <scene.Component model={scene} />;
};
