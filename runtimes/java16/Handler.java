public class Handler {
    public static byte[] Handle(byte[] msg) throws Exception {
        var x = "hi " + new String(msg);
        return x.getBytes();
    }
}
