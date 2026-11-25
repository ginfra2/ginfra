/* generated using openapi-typescript-codegen -- do no edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */

import type { ConfigSetItem } from './ConfigSetItem';

export type ConfigServiceSetRequest = {
    failmode?: ConfigServiceSetRequest.failmode;
    items?: Array<ConfigSetItem>;
};

export namespace ConfigServiceSetRequest {

    export enum failmode {
        NONE = 'none',
        ROLLBACK = 'rollback',
    }


}
