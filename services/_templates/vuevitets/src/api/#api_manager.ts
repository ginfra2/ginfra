// General API stuff

import {CancelablePromise, DefaultService, Giconfig, type Specifics} from "@/api/giconfig";
import UserFetch from '../components/UserFetch.vue'
import {Giuser} from "@/api/giuser";

export interface ApiClientInfo {
    initializing:   boolean;

    configurl:      string;
    accesstoken:    string;
    deployment:     string;

    configClient?:  Giconfig;

    serverFault?:        string;

    // Add actual AUTH later.
}

export let LiveApiClientInfo : ApiClientInfo

export async function InitApiClient() {

    return new Promise<void>((resolve, reject) => {
        LiveApiClientInfo = {
            initializing: true,
            configurl: import.meta.env.VITE_GINFRA_CONFIG_URI,
            accesstoken: import.meta.env.VITE_GINFRA_ACCESS_TOKEN,
            deployment: import.meta.env.VITE_GINFRA_DEPLOYMENT,
            serverFault: ""
        }
        if ((LiveApiClientInfo.configurl == undefined) || (LiveApiClientInfo.configurl == "") ||
            (LiveApiClientInfo.accesstoken == undefined) || (LiveApiClientInfo.accesstoken == "") ||
            (LiveApiClientInfo.deployment == undefined) || (LiveApiClientInfo.deployment == "")) {
            throw Error("Configuration provider access properties not complete.")
        }
        LiveApiClientInfo.configClient = new Giconfig({
            BASE: LiveApiClientInfo.configurl,
        });
        console.log("Start configuration.")
        doConfig(LiveApiClientInfo.configClient)
            .then(() => {
                console.log("Configuration complete.")
            })
            .catch((err) => {
                console.log("Configuration failed.")
                LiveApiClientInfo.serverFault = err.toString()
            })
            .finally(() => {
                resolve()
            })
    })

}

async function doConfig(configClient: Giconfig) : Promise<void> {
    return new Promise<void>((resolve, reject) => {

        configClient.default.giGetSpecifics()
            .then((specifics) => {

                // Do more stuff here.

            })
            .catch((err) => {
                reject("Configuration service unavailable.  " + err.toString())
            })
    })
}
