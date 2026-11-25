/*

 */
package tech.ginfra.service.controller;

import micronaut.api.DefaultApi;
import micronaut.model.*;
import io.micronaut.scheduling.TaskExecutors;
import io.micronaut.scheduling.annotation.ExecuteOn;
import tech.ginfra.Ginfra;

@io.micronaut.http.annotation.Controller
public class Controller implements DefaultApi {


    public Controller() {
    }

    @ExecuteOn(TaskExecutors.BLOCKING)
    public ControlDatasourceInitResponse giControlDatasourceInit(ControlDatasourceInitRequest controlDatasourceInitRequest) {
        throw new NotFoundException("Control not implemented.");
    }

    @ExecuteOn(TaskExecutors.BLOCKING)
    public ControlActionResponse giControlManageReset(String auth) {
        throw new NotFoundException("Control not implemented.");
    }

    @ExecuteOn(TaskExecutors.BLOCKING)
    public ControlActionResponse giControlManageStop(String auth) {
        throw new NotFoundException("Control not implemented.");
    }

}

