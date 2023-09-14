import * as React from 'react';
import { Redirect, Route, Switch } from 'react-router-dom';
import { HomePage } from '../../pages/Home';
import { prefixRoute } from '../../utils/utils.routing';
import { ROUTES } from '../../constants';

import { CausePluginPage } from '../../pages/Cause';
import { TopologyPluginPage } from '../../pages/Topology';
import { TracePluginPage } from '../../pages/TraceProfile';

export const Routes = () => {
  return (
    <Switch>
      <Route path={prefixRoute(`${ROUTES.Home}`)} component={HomePage} />
      <Route path={prefixRoute(`${ROUTES.Cause}`)} component={CausePluginPage} />
      <Route path={prefixRoute(`${ROUTES.Topology}`)} component={TopologyPluginPage} />
      <Route path={prefixRoute(`${ROUTES.Trace}`)} component={TracePluginPage} />
      <Redirect to={prefixRoute(ROUTES.Home)} />
    </Switch>
  );
};
