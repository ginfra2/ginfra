package tech.ginfra.service_gipet.exception;

public class NotFoundException extends RuntimeException {
    public String message;

    public NotFoundException() {
        message = "Not found";
    }

    public NotFoundException(String msg) {
        message = msg;
    }
}
