/*
Package config
Copyright (c) 2023 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

This file defines common naming for configuration points.  You will notice that match the points in the example
config json in config.go.  Since ginfra apps and the services rely on a consistent configuration scheme, these should
be used when possible.  Also these are all const not because I hope the compiler would be smart enough to have only
one instance of each string, but if it isn't I can change them to vars.
*/
package config

// #####################################################################################################################
// # NAMING FOR PARAMETER NAMING AND/OR COMMAND LINE

const CL_BUNDLE = "bundle"
const CL_CLEAR = "clear"
const CL_COMMAND_LINE = "cmd"
const CL_DONT_START = "dont_start"
const CL_SPECIFICS = "specifics"
const CL_DEBUG = "debug"
const CL_DEBUGGER = "debugger"
const CL_DEPLOYMENT = "deployment"
const CL_SERVICE = "service"
const CL_SERVICE_HOME = "service_home"
const CL_INIT_CONTAINER = "init_container"
const CL_INSTANCE_TYPE = "type"
const CL_NAME = "name"
const CL_NO_ROLLBACK = "no_rollback"
const CL_HOST = "host"
const CL_THIS = "this"
const CL_FQDN = "fqdn"
const CL_PORT = "port"
const CL_PORT_SUFFIX = "_port"
const CL_HOSTDIR = "host_dir"
const CL_CLASS = "class"
const CL_PATH = "path"
const CL_INTERFACEDIR = "interfacedir"
const CL_TEMPLATES = "templates"
const CL_PORTAL = "portal"
const CL_NO_CONTROL = "no_control"
const CL_GPUS = "gpus"
const CL_AUTH = "auth"
const CL_MODEL = "model"

// #####################################################################################################################
// # NAMING FOR NODES IN CONFIGURATION TREE

// Common nodes

const CR_AUTH = "auth"
const CR_SERVICE_AUTH = "service_auth"
const CR_TIME = "time"

// Root nodes
const CR_DEPLOYMENTS = "deployments"
const CR_LIVE = "live"
const CR_RESERVATIONS = "reservations"

const CR_DEPLOYMENT_DEFAULT = "default"

// Deployment nodes
const CRD_SERVICES = "services"
const CRD_PORTALS = "portals"

// Reservations
const CR_RESERVATIONS_PORTS = "ports"

// Deployment -> Values
const CRD_SERVICES__VALUES = "values"

// Deployment -> Services nodes
const CRD_SERVICES__SERVICE = "service"
const CRD_SSERVICE__CONFIG = "config"
const CRD_SERVICES__AUTH = "auth"
const CRD_SERVICES__SERVICE_AUTH = "service_auth"
const CRD_SSERVICE__CONFIG__PORT = "port"
const CRD_SSERVICE__CONFIG__IMAGE = "image"
const CRD_SSERVICE__CONFIG__CONFIG_PROVIDER = "config_provider"
const CRD_SSERVICE__CONFIG__SERVICE_CLASS = "service_class"
const CRD_SSERVICE__STATE = "state"
const CRD_SSERVICE__CONFIG__SPECIFICS = "specifics"
const CRD_SSERVICE__CONFIG__SPECIFICS__SRVKEY = "server.key"
const CRD_SSERVICE__CONFIG__SPECIFICS__SRVCERT = "server.cert"

const CRD_SERVICES__MANAGEMENT = "management"
const CRD_MSERVICE__INSTANCE = "instance"
const CRD_MSERVICE__INSTANCE__TYPE = CL_INSTANCE_TYPE
const CRD_MSERVICE__INSTANCE__ID = "id"
const CRD_MSERVICE__INSTANCE__IP = "ip"
const CRD_MSERVICE__INSTANCE__FQDN = "fqdn"

const CRD_SERVICES_INITID = "initid"

// Deployment -> Discovery nodes
const CRD_DISCOVERY = "discovery"

// #####################################################################################################################
// # NAMING FOR VALUES

const CN_INSTANCE_TYPE__LOCAL = "local"
const CN_INSTANCE_TYPE__HOST = "host"
const CN_INSTANCE_TYPE__MANAGED = "managed"
const CN_INSTANCE_TYPE__KUBE = "kube"

const CN_RESERVATION_TYPE__PORT = "port"

const CN_DEPLOYMENT = "deployment"

// #####################################################################################################################
// # ODDS

const CO_ANY_PORTAL = "*"
