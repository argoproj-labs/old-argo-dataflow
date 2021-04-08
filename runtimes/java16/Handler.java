public class Handler {
    public static byte[][] Handle(byte[] msg) throws Exception {
        var x = "hi " + new String(msg);
        return new byte[][]{x.getBytes()};
    }
}
