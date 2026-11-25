/* generated using openapi-typescript-codegen -- do no edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */

import type { namevalue } from './namevalue';

export type ConfigServiceRegistrationRequest = {
    /**
     * URI to configuration provider.  This is not authoritive.
     */
    configprovider?: string;
    /**
     * Deployment name.  If blank it will be the default deployment.
     */
    deployment?: string;
    image?: string;
    /**
     * System unique name
     */
    name?: string;
    /**
     * Exposed service port
     */
    port?: number;
    serviceclass?: string;
    specifics?: Array<namevalue>;
};
