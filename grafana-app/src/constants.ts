import pluginJson from './plugin.json';

export const PLUGIN_BASE_URL = `/a/${pluginJson.id}`;

export enum ROUTES {
  Home = 'home',
  WithTabs = 'page-with-tabs',
  WithDrilldown = 'page-with-drilldown',
  HelloWorld = 'hello-world',
  Cause = 'cause',
  Topology = 'topology',
  Trace = 'trace'
}

export const PrometheusName = 'Prometheus';
export const DATASOURCE_REF = {
  uid: 'a4f53ab2-c01b-4472-bea2-cb07e228e7f8',
  type: 'Prometheus',
};
