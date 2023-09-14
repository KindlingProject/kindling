import React from 'react';
import { getScene } from './helloWorldScene';

export const HelloWorldPluginPage = () => {
  const scene = getScene();

  return <scene.Component model={scene} />;
};
