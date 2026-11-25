/* generated using openapi-typescript-codegen -- do no edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { authtoken } from '../models/authtoken';
import type { ConfigGetInitResponse } from '../models/ConfigGetInitResponse';
import type { ConfigGetStateResponse } from '../models/ConfigGetStateResponse';
import type { ConfigPath } from '../models/ConfigPath';
import type { ConfigPortalRequest } from '../models/ConfigPortalRequest';
import type { ConfigQueryResponse } from '../models/ConfigQueryResponse';
import type { ConfigServiceInstanceRequest } from '../models/ConfigServiceInstanceRequest';
import type { ConfigServiceRegistrationRequest } from '../models/ConfigServiceRegistrationRequest';
import type { ConfigServiceRegistrationResponse } from '../models/ConfigServiceRegistrationResponse';
import type { ConfigServiceSetRequest } from '../models/ConfigServiceSetRequest';
import type { ConfigServiceSetStateRequest } from '../models/ConfigServiceSetStateRequest';
import type { ConfigSetStateResponse } from '../models/ConfigSetStateResponse';
import type { DiscoverPortalResponse } from '../models/DiscoverPortalResponse';
import type { DiscoverServiceResponse } from '../models/DiscoverServiceResponse';
import type { Specifics } from '../models/Specifics';

import type { CancelablePromise } from '../core/CancelablePromise';
import type { BaseHttpRequest } from '../core/BaseHttpRequest';

export class DefaultService {

    constructor(public readonly httpRequest: BaseHttpRequest) {}

    /**
     * Deregister a service.
     * A priviledged deregistration request.
     * @param confauth Administrative access token given to the config service when started.
     * @param name Unique name of the service.
     * @param deployment Unique deployment name.
     * @returns ConfigServiceRegistrationResponse Success
     * @throws ApiError
     */
    public giConfigDeregisterService(
confauth: authtoken,
name: string,
deployment: string,
): CancelablePromise<ConfigServiceRegistrationResponse> {
        return this.httpRequest.request({
            method: 'POST',
            url: '/giconfig/deregister/{deployment}/{name}/',
            path: {
                'name': name,
                'deployment': deployment,
            },
            headers: {
                'confauth': confauth,
            },
            errors: {
                400: `Bad request`,
                403: `Not authorized`,
                500: `InternalError`,
            },
        });
    }

    /**
     * Discover a portal for a deployment.
     * Discover a portal location for a deployment.  A portal provides external access to a service.
     * @param deployment Deployment name.
     * @param serviceclass Service class name.
     * @param confauth Access token.
     * @returns DiscoverPortalResponse Success
     * @throws ApiError
     */
    public giDiscoverPortal(
deployment: string,
serviceclass: string,
confauth: authtoken,
): CancelablePromise<DiscoverPortalResponse> {
        return this.httpRequest.request({
            method: 'GET',
            url: '/giconfig/discover/portal/{deployment}/{serviceclass}',
            path: {
                'deployment': deployment,
                'serviceclass': serviceclass,
            },
            headers: {
                'confauth': confauth,
            },
            errors: {
                400: `Bad request`,
                404: `Not found`,
                500: `InternalError`,
            },
        });
    }

    /**
     * Discover a service.
     * Discover a service entrypoint uri.
     * @param deployment Deployment name.
     * @param serviceclass Service class name.
     * @param confauth Access token.
     * @returns DiscoverServiceResponse Success
     * @throws ApiError
     */
    public giDiscoverService(
deployment: string,
serviceclass: string,
confauth: authtoken,
): CancelablePromise<DiscoverServiceResponse> {
        return this.httpRequest.request({
            method: 'GET',
            url: '/giconfig/discover/service/{deployment}/{serviceclass}',
            path: {
                'deployment': deployment,
                'serviceclass': serviceclass,
            },
            headers: {
                'confauth': confauth,
            },
            errors: {
                400: `Bad request`,
                404: `Not found`,
                500: `InternalError`,
            },
        });
    }

    /**
     * Register or update a service instance.
     * A priviledged registration request.
     * @param confauth Administrative access token given to the config service when started.
     * @param requestBody Register a service instance.
     * @returns ConfigServiceRegistrationResponse Success
     * @throws ApiError
     */
    public giConfigRegisterInstance(
confauth: authtoken,
requestBody: ConfigServiceInstanceRequest,
): CancelablePromise<ConfigServiceRegistrationResponse> {
        return this.httpRequest.request({
            method: 'POST',
            url: '/giconfig/instance',
            headers: {
                'confauth': confauth,
            },
            body: requestBody,
            mediaType: 'application/json',
            errors: {
                400: `Bad request`,
                403: `Not authorized`,
                500: `InternalError`,
            },
        });
    }

    /**
     * Register or update a portal.
     * Register or update a portal.
     * @param confauth Administrative access token given to the config service when started.
     * @param operation Unique deployment name.
     * @param deployment Unique deployment name.
     * @param requestBody Register a portal.
     * @returns any Success
     * @throws ApiError
     */
    public giConfigPortal(
confauth: authtoken,
operation: 'put' | 'remove',
deployment: string,
requestBody: ConfigPortalRequest,
): CancelablePromise<any> {
        return this.httpRequest.request({
            method: 'POST',
            url: '/giconfig/portal/{operation}/{deployment}/',
            path: {
                'operation': operation,
                'deployment': deployment,
            },
            headers: {
                'confauth': confauth,
            },
            body: requestBody,
            mediaType: 'application/json',
            errors: {
                400: `Bad request`,
                403: `Not authorized`,
                500: `InternalError`,
            },
        });
    }

    /**
     * Query config.
     * Query configuration data.
     * @param qpath The dot separated path the configuration point.
     * @param confauth Access token.
     * @returns ConfigQueryResponse Success
     * @throws ApiError
     */
    public giConfigQueryConfig(
qpath: ConfigPath,
confauth: authtoken,
): CancelablePromise<ConfigQueryResponse> {
        return this.httpRequest.request({
            method: 'GET',
            url: '/giconfig/query',
            headers: {
                'qpath': qpath,
                'confauth': confauth,
            },
            errors: {
                400: `Bad request`,
                404: `Not found`,
                500: `InternalError`,
            },
        });
    }

    /**
     * Register a service.
     * A privileged registration request.
     * @param confauth Administrative access token given to the config service when started.
     * @param requestBody Register a service.
     * @returns ConfigServiceRegistrationResponse Success
     * @throws ApiError
     */
    public giConfigRegisterService(
confauth: authtoken,
requestBody: ConfigServiceRegistrationRequest,
): CancelablePromise<ConfigServiceRegistrationResponse> {
        return this.httpRequest.request({
            method: 'POST',
            url: '/giconfig/register',
            headers: {
                'confauth': confauth,
            },
            body: requestBody,
            mediaType: 'application/json',
            errors: {
                400: `Bad request`,
                403: `Not authorized`,
                500: `InternalError`,
            },
        });
    }

    /**
     * Remove config.
     * Remove configuration data.
     * @param qpath The dot separated path the configuration point.
     * @param confauth Access token.
     * @returns any Success
     * @throws ApiError
     */
    public giRemoveConfig(
qpath: ConfigPath,
confauth: authtoken,
): CancelablePromise<any> {
        return this.httpRequest.request({
            method: 'GET',
            url: '/giconfig/remove',
            headers: {
                'qpath': qpath,
                'confauth': confauth,
            },
            errors: {
                400: `Bad request`,
                404: `Not found`,
                500: `InternalError`,
            },
        });
    }

    /**
     * Set configurations.
     * A privileged registration request.
     * @param confauth Access token.
     * @param requestBody Set config.
     * @returns any Success
     * @throws ApiError
     */
    public giConfigSetConfig(
confauth: authtoken,
requestBody: ConfigServiceSetRequest,
): CancelablePromise<any> {
        return this.httpRequest.request({
            method: 'POST',
            url: '/giconfig/set',
            headers: {
                'confauth': confauth,
            },
            body: requestBody,
            mediaType: 'application/json',
            errors: {
                400: `Bad request`,
                403: `Not authorized`,
                500: `InternalError`,
            },
        });
    }

    /**
     * Get the current service state.
     * An service state request.
     * @param confauth Service access token.
     * @param deployment Unique deployment name.
     * @param name Unique name of the service.
     * @returns ConfigGetStateResponse Success
     * @throws ApiError
     */
    public giConfigGetState(
confauth: authtoken,
deployment: string,
name: string,
): CancelablePromise<ConfigGetStateResponse> {
        return this.httpRequest.request({
            method: 'GET',
            url: '/giconfig/state/get/{deployment}/{name}/',
            path: {
                'deployment': deployment,
                'name': name,
            },
            headers: {
                'confauth': confauth,
            },
            errors: {
                400: `Bad request`,
                403: `Not authorized`,
                500: `InternalError`,
            },
        });
    }

    /**
     * Get the current initializaton state.
     * An initializaton state request.
     * @param confauth Service access token.
     * @param deployment Unique deployment name.
     * @param name Unique name of the service.
     * @param id Unique idenity of the calling instance.
     * @returns ConfigGetInitResponse Success
     * @throws ApiError
     */
    public giConfigGetInit(
confauth: authtoken,
deployment: string,
name: string,
id: string,
): CancelablePromise<ConfigGetInitResponse> {
        return this.httpRequest.request({
            method: 'GET',
            url: '/giconfig/state/init/{deployment}/{name}/{id}/',
            path: {
                'deployment': deployment,
                'name': name,
                'id': id,
            },
            headers: {
                'confauth': confauth,
            },
            errors: {
                400: `Bad request`,
                403: `Not authorized`,
                500: `InternalError`,
            },
        });
    }

    /**
     * Set State.
     * A set state request.
     * @param confauth Access token.
     * @param deployment Unique deployment name.
     * @param name Unique name of the service.
     * @param requestBody Set config.
     * @returns ConfigSetStateResponse Success
     * @throws ApiError
     */
    public giConfigSetState(
confauth: authtoken,
deployment: string,
name: string,
requestBody: ConfigServiceSetStateRequest,
): CancelablePromise<ConfigSetStateResponse> {
        return this.httpRequest.request({
            method: 'POST',
            url: '/giconfig/state/set/{deployment}/{name}',
            path: {
                'deployment': deployment,
                'name': name,
            },
            headers: {
                'confauth': confauth,
            },
            body: requestBody,
            mediaType: 'application/json',
            errors: {
                400: `Bad request`,
                403: `Not authorized`,
                500: `InternalError`,
            },
        });
    }

    /**
     * Get specifics for this service
     * Get specifics
     * @returns Specifics Success
     * @throws ApiError
     */
    public giGetSpecifics(): CancelablePromise<Specifics> {
        return this.httpRequest.request({
            method: 'GET',
            url: '/specifics',
            errors: {
                500: `InternalError`,
            },
        });
    }

}
