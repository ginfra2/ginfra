/*
Nodejs service
// >A#

You may license this file however you wish.

// >A#
*/
const express = require('express');
const path = require('path');
const cors = require('cors');
const fs = require('fs');
const https = require('https');
const ginfra = require('./ginfra.js');
const winston = require('winston');
const fetch = require('node-fetch').default;
const {text} = require("express");

// # YOUR IMPORTS START HERE
// >B###############################################################################################################

// CORS configuration -- update with your headers
const corsOptions = {
    methods: ['GET', 'POST'],
    allowedHeaders: ['Content-Type']
};

// >B###############################################################################################################
// #  YOUR IMPORTS END

const app = express();

// ---------------------------------------------------------------------------------------------------------------------
// -- CONFIGURATION ----------------------------------------------------------------------------------------------------

const logger = winston.createLogger({
    transports: [
        new winston.transports.File({ filename: 'log/server.log' })
    ]
});

let web_host = ginfra.getServiceSpecificsConfig('web_host');
if (!web_host) {
    web_host = '0.0.0.0';
    logger.info('No host specified, using default');
}
logger.info('Web Host: ', web_host);

let web_port = ginfra.getServiceSpecificsConfig('web_port');
if (!web_port) {
    web_port = '8888';
    logger.info('No port specified, using default');
}
logger.info('Web Port: ', web_port);

let service_port = ginfra.getConfig('port');
if (!service_port) {
    service_port = '8900';
    logger.info('No service port specified, using default');
}
logger.info('Service Port: ', web_port);


// # YOUR CONFIGURATIONS START HERE
// >C##############################################################################################################


// >C###############################################################################################################
// #  YOUR CONFIGURATIONS END

// ---------------------------------------------------------------------------------------------------------------------
// -- PIPELINE ----------------------------------------------------------------------------------------------------

app.use(cors(corsOptions));
app.use(express.json());
app.use(express.static(path.join(__dirname, 'public')));

app.use((req, res, next) => {
    logger.info(`${new Date().toISOString()} - ${req.method} ${req.url}`);
    // TODO add debugging mode that logs request data.
    next();
});

app.use((err, req, res, next) => {
    logger.error(err.stack);
    res.status(500).json({ error: 'Internal Server Error' });
});

// ---------------------------------------------------------------------------------------------------------------------
// -- SERVER ---------------------------------------------------------------------------------------------------------

let server;
if (fs.existsSync('server.key') && fs.existsSync('server.cert')) {
    const privateKey = fs.readFileSync('server.key', 'utf8');
    const certificate = fs.readFileSync('server.cert', 'utf8');
    const credentials = {key: privateKey, cert: certificate};

    server = https.createServer(credentials, app);
    server.listen(web_port, web_host, () => {
        logger.info(`Server running at https://${web_host}:${web_port}`);
    });
} else {
    server = app.listen(web_port, web_host, () => {
        logger.info(`Server running at http://${web_host}:${web_port}`);
    });
}

// Graceful shutdown
process.on('SIGTERM', () => {
    server.close(() => {
        logger.info('Server closed');
    });
});

// ---------------------------------------------------------------------------------------------------------------------
// -- HANDLERS ---------------------------------------------------------------------------------------------------------

// ^^###################################################################################################################
// # ADDITIONAL CODE GOES BELOW HERE  PUT YOUR HANDLERS HERE

