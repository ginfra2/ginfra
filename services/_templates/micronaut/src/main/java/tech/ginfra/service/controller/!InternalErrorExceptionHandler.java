package tech.ginfra.service.controller;

import io.micronaut.context.annotation.Requires;
import io.micronaut.http.HttpRequest;
import io.micronaut.http.HttpResponse;
import io.micronaut.http.server.exceptions.ExceptionHandler;
import jakarta.inject.Singleton;
import tech.ginfra.api.service_giconfig.client.model.ErrorResponse;
import tech.ginfra.service.exception.InternalErrorException;

@Singleton
@Requires(classes = InternalErrorException.class)
public class InternalErrorExceptionHandler implements ExceptionHandler<InternalErrorException,
        HttpResponse<ErrorResponse>> {

    @Override
    public HttpResponse<ErrorResponse> handle(HttpRequest request, InternalErrorException exception) {
        var err = new ErrorResponse();
        err.message(exception.message);
        err.diagnostics(exception.diagnostic);
        return HttpResponse.badRequest(err);
    }
}