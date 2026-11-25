package tech.ginfra.service.controller;

import io.micronaut.context.annotation.Requires;
import io.micronaut.http.HttpRequest;
import io.micronaut.http.HttpResponse;
import io.micronaut.http.server.exceptions.ExceptionHandler;
import jakarta.inject.Singleton;
import tech.ginfra.api.service_giconfig.client.model.ErrorResponse;
import tech.ginfra.service.exception.BadRequestException;

@Singleton
@Requires(classes = BadRequestException.class)
public class BadRequestExceptionHandler implements ExceptionHandler<BadRequestException,
        HttpResponse<ErrorResponse>> {

    @Override
    public HttpResponse<ErrorResponse> handle(HttpRequest request, BadRequestException exception) {
        var err = new ErrorResponse();
        err.message(exception.message);
        return HttpResponse.badRequest(err);
    }
}