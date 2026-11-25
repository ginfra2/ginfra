package tech.ginfra.service.exception;

public class InternalErrorException extends RuntimeException {
    public String message;
    public String diagnostic;

    public InternalErrorException() {
        message = "Internal Server Error";
    }
    public InternalErrorException(String msg, String diag) {
        message = msg;
        diagnostic = diag;
    }
}
