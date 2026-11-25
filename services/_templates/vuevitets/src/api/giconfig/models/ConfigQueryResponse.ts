/* generated using openapi-typescript-codegen -- do no edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */

import type { base64 } from './base64';
import type { ConfigPath } from './ConfigPath';

export type ConfigQueryResponse = {
    /**
     * Base64 json block representing the config.
     */
    block?: base64;
    qpath?: ConfigPath;
};
