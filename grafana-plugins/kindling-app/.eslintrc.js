module.exports = {
  extends: ['@grafana/eslint-config', 'plugin:react-hooks/recommended'],
  rules: {
    'react/prop-types': 'off',
    'react-hooks/exhaustive-deps': 'error',
  },
};
