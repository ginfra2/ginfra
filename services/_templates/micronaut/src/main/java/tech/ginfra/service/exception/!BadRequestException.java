package tech.ginfra.service.exception;

public class BadRequestException extends RuntimeException {
    public String message;

    public BadRequestException() {
        message = "Bad request";
    }
    public BadRequestException(String m) {
        message = m;
    }
}
