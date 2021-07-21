import java.util.Map;

public class Handler {
    public static byte[] Handle(byte[] msg, Map<String, String> context) throws Exception {
        return ("hi! " + new String(msg)).getBytes("UTF-8");
    }
}