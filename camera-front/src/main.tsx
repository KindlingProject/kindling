import React, { useEffect } from 'react'
import ReactDOM from 'react-dom/client'
import routes from './router';
import { HashRouter as Router } from 'react-router-dom';
import { Provider } from 'react-redux';
import { PersistGate } from 'redux-persist/integration/react';
import { ConfigProvider } from 'antd';
import zhCN from 'antd/lib/locale/zh_CN';
import { persistor, store } from './store/storeConfig';

import './styles/index.less'
import './styles/common.less';
import './styles/theme/theme-color.less';

const RootBody = () => {

  useEffect(() => {
    // window['eventListenerList'] = [];
    let body = document.getElementsByTagName('body')[0];
    body.className = `light-theme`;
  }, []);

  return (
      <ConfigProvider locale={zhCN}>
          <Provider store={store}>
              <PersistGate persistor={persistor}>
                  <Router>{routes}</Router>
              </PersistGate>
          </Provider>
      </ConfigProvider>
      // <React.StrictMode>
      //     <Router>{routes}</Router>
      // </React.StrictMode>
  );
};

ReactDOM.createRoot(document.getElementById('root')!).render(
  // <React.StrictMode>
    <RootBody />
  // </React.StrictMode>
)

