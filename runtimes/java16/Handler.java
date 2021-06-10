import java.util.Map;

public class Handler {
    public static byte[] Handle(byte[] msg, Map<String,String> context) throws Exception {
        var x = "hi " + new String(msg);
        return x.getBytes();
    }
}
