import { SceneQueryRunner, SceneTimeRange, EmbeddedScene, SceneFlexLayout, SceneFlexItem, PanelBuilders } from '@grafana/scenes';

export function getScene() {
  const queryRunner = new SceneQueryRunner({
    datasource: {
      uid: 'a4f53ab2-c01b-4472-bea2-cb07e228e7f8',
      type: 'Prometheus'
    },
    queries: [
      {
        refId: 'A',
        expr: 'rate(prometheus_http_requests_total{}[5m])',
      },
    ],
    $timeRange: new SceneTimeRange({ from: 'now-5m', to: 'now' }),
  });

  return new EmbeddedScene({
    $data: queryRunner,
    body: new SceneFlexLayout({
      children: [
        new SceneFlexItem({
          width: '100%',
          height: 300,
          // body: PanelBuilders.text().setTitle('Hello world panel').setOption('content', 'Hello world!').build(),
          body: PanelBuilders.timeseries().setTitle('test').build(),
        }),
      ],
    }),
  });
}
