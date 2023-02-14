'use strict';

const express = require('express');
const app = express();
const http = require('http');
const serveStatic = require('serve-static');
const bodyParser = require('body-parser');
const _ = require('lodash');
const async = require('async');
const path = require('path');
const compression = require('compression');
const rateLimit = require('express-rate-limit');
const history = require('connect-history-api-fallback');
const { createProxyMiddleware } = require('http-proxy-middleware');

const settings = require('./settings');
const port = settings.port;

const restream = function (proxyReq, req) {
    let masterIp = req.headers.authorization;
    proxyReq.setHeader('Authorization', masterIp || '');
};
const esserverAdress = 'http://' + settings.apmServerConfig.host + ':' + settings.apmServerConfig.port;
const esserverProxy = createProxyMiddleware('/camera', {
    target: esserverAdress,
    secure: false,
    changeOrigin: true,
    // onProxyReq: restream,
    // onProxyRes: async function (proxyRes, req, res) {
    //     // console.log(proxyRes, 'proxyRes');
    //     if (typeof req.session.username === 'undefined') {
    //         req.session.destroy();
    //         res.status(401).json({
    //             msg: noAuthorizeMsg,
    //         });
    //     }
    // },
});
app.use(esserverProxy);
const profileAddress = 'http://' + settings.profileConfig.host + ':' + settings.profileConfig.port;
const profileProxy = createProxyMiddleware('/profile', {
    target: profileAddress,
    secure: false,
    changeOrigin: true,
});
app.use(profileProxy);

const limiter = rateLimit({
    windowMs: 10 * 60 * 1000, // 1 minute
    max: 100
});
app.use(limiter);

app.use(
    bodyParser.urlencoded({
        extended: true,
    })
);
app.use(bodyParser.json());
app.use(compression());
// app.use(history());

app.use(express.static('../dist'));
console.log(__dirname);
// app.set('views', path.join(__dirname, 'dist'));
// app.use(serveStatic(path.join(__dirname, 'dist')));

app.get("/test", function(req, res, next) {
    // res.render("test.html");
    res.status(200).json({
        test: true
    });
});

app.use('/file', require('./routers/file.js'));

var httpServer = http.createServer(app);
// var httpsServer = https.createServer(credentials, app);
httpServer.listen(port, function() {
    console.log('Express http server listening on port ' + port);
});