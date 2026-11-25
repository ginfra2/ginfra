package tech.ginfra.service.controller;

import jakarta.inject.Singleton;
import tech.ginfra.service.exception.NotAuthorizedException;

import io.micronaut.context.annotation.Requires;
import io.micronaut.http.HttpRequest;
import io.micronaut.http.HttpResponse;
import io.micronaut.http.server.exceptions.ExceptionHandler;

@Singleton
@Requires(classes = NotAuthorizedException.class)
public class NotAuthorizedExceptionHandler implements ExceptionHandler<NotAuthorizedException,
        HttpResponse<String>> {

    @Override
    public HttpResponse<String> handle(HttpRequest request, NotAuthorizedException exception) {
        return HttpResponse.unauthorized();
    }
}