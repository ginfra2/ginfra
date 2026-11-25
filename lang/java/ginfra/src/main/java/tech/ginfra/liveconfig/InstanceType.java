package tech.ginfra.liveconfig;

public enum InstanceType {
    ITBad(0),
    ITLocal(1),
    ITHost(2),
    ITManaged(3);

    private final int value;
    InstanceType(int v) {
        value = v;
    }
    public int getValue() { return value; }
    public static InstanceType getInstanceType(int i) {
        var r = ITBad;
        try {
            r = InstanceType.values()[i];
        } catch (Exception e) {
            // Don't care, it's bad.
        }
        return r;
    }
}
