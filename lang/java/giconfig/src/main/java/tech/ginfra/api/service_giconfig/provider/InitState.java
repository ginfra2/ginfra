package tech.ginfra.api.service_giconfig.provider;

public enum InitState {
    InitNew(0),
    InitPending(1),
    InitDone(2);

    private final int value;
    private InitState(int v) {
        value = v;
    }
    public int getValue() { return value; }


}
