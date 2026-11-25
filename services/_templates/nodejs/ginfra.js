const fs = require('fs');

let serviceConfig = {};
let configContent = "";

try {
    if (fs.existsSync('service_config.json')) {
        configContent = fs.readFileSync('service_config.json', 'utf-8');
    } else if (fs.existsSync('../service_config.json')) {
        configContent = fs.readFileSync('../service_config.json', 'utf-8');
    } else if (fs.existsSync('../service/service_config.json')) {
        configContent = fs.readFileSync('../service/service_config.json', 'utf-8');
    } else if (fs.existsSync('.service/service_config.json')) {
        configContent = fs.readFileSync('service/service_config.json', 'utf-8');
    } else {
        throw new Error('service_config.json file not found in any of the expected locations');
    }
    serviceConfig = JSON.parse(configContent);
} catch (error) {
    console.error('Error loading service_config.json:', error);
    serviceConfig = {};
}

function getConfig(key) {
    return serviceConfig[key];
}

function getServiceSpecificsConfig(key) {
    return serviceConfig["specifics"][key];
}


module.exports = {
    getConfig,
    getServiceSpecificsConfig
};

