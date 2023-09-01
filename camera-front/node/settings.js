const setting = {
    esServerConfig: {
        host: 'localhost',
        port: '9200',
        trace_index: 'single_net_request_metric_group_dev',
        onoffcpu_index: 'camera_event_group_dev'
    },
    serverConfig: {
        host: '10.10.103.96',
        port: '8081',
    },
    profileConfig: {
        host: 'localhost',
        port: '9503',
    },
    traceFilePath: '/tmp/kindling',
    ratelimit: {
        windowMs: 10 * 60 * 1000,
        max: 500
    },
    port: 9900,
};
module.exports = setting;
