/* generated using openapi-typescript-codegen -- do no edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */

export type ConfigServiceInstanceRequest = {
    /**
     * Deployment name.  If blank it will be the default deployment.
     */
    deployment?: string;
    /**
     * FQDN for the entrypoint for the service.
     */
    fqdn?: string;
    /**
     * IP address.
     */
    id?: string;
    /**
     * IP address.
     */
    ip?: string;
    /**
     * System unique name
     */
    name?: string;
    /**
     * Service friendly config authorization token.
     */
    sauth?: string;
    /**
     * Deployment type.
     */
    type?: string;
};
