const setting = {
    apmServerConfig: {
        host: 'localhost',
        port: '2234',
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
    port: 9504,
};
module.exports = setting;
