package tech.ginfra.service.controller;

import io.micronaut.context.annotation.Requires;
import io.micronaut.http.HttpRequest;
import io.micronaut.http.HttpResponse;
import io.micronaut.http.HttpStatus;
import io.micronaut.http.server.exceptions.ExceptionHandler;
import jakarta.inject.Singleton;
import tech.ginfra.service.exception.NotFoundException;

@Singleton
@Requires(classes = NotFoundException.class)
public class NotFoundExceptionHandler implements ExceptionHandler<NotFoundException,
        HttpResponse<String>> {

    @Override
    public HttpResponse<String> handle(HttpRequest request, NotFoundException exception) {
        return HttpResponse.status(HttpStatus.NOT_FOUND, exception.message);
    }
}