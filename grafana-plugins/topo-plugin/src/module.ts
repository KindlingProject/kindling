import { PanelPlugin } from '@grafana/data';
import { SimpleOptions } from './types';
import { TopologyPanel } from './topologyPanel';

export const plugin = new PanelPlugin<SimpleOptions>(TopologyPanel).setPanelOptions(builder => {
  return builder
    // .addTextInput({
    //   path: 'text',
    //   name: 'Simple text option',
    //   description: 'Description of panel option',
    //   defaultValue: 'Default value of text input option',
    // })
    // .addRadio({
    //   path: 'seriesCountSize',
    //   defaultValue: 'sm',
    //   name: 'Normal',
    //   settings: {
    //     options: [
    //       {
    //         value: 'sm',
    //         label: 'Small',
    //       },
    //       {
    //         value: 'md',
    //         label: 'Medium',
    //       },
    //       {
    //         value: 'lg',
    //         label: 'Large',
    //       },
    //     ],
    //   },
    //   showIf: config => config.showLatency,
    // })
    .addBooleanSwitch({
      path: 'showLatency',
      name: 'Latency Config',
      description: 'The value change latency threshold config',
      defaultValue: false,
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
});
