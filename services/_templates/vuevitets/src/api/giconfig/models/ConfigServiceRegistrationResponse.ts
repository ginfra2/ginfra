/* generated using openapi-typescript-codegen -- do no edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */

import type { authtoken } from './authtoken';

export type ConfigServiceRegistrationResponse = {
    /**
     * Auth token the service should use to access this service.
     */
    confauth?: authtoken;
    /**
     * System unique name
     */
    name?: string;
};
