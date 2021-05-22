import com.sun.net.httpserver.HttpServer;

import java.io.ByteArrayOutputStream;
import java.io.InputStreamReader;
import java.net.InetSocketAddress;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.nio.charset.StandardCharsets;
import java.util.concurrent.Executors;

public class Main {
    public static void main(String[] args) throws Exception {
        var server = HttpServer.create(new InetSocketAddress("localhost", 8080), 0);
        server.setExecutor(Executors.newSingleThreadExecutor());
        server.createContext("/ready", he -> he.sendResponseHeaders(204, 0));
        server.createContext("/messages", he -> {
            // read all input bytes
            var isr = new InputStreamReader(he.getRequestBody(), StandardCharsets.UTF_8);
            var b = 0;
            var in = new ByteArrayOutputStream();
            while ((b = isr.read()) != -1) {
                in.write((char) b);
            }
            isr.close();
            try {
                var out = Handler.Handle(in.toByteArray());
                if (out != null) {
                    he.sendResponseHeaders(201, 0);
                    try (var os = he.getResponseBody()) {
                        os.write(out);
                    }
                } else{
                    he.sendResponseHeaders(204, 0);
                }
            } catch (Exception e) {
                he.sendResponseHeaders(500, 0);
                try (var os = he.getResponseBody()) {
                    os.write(e.getMessage().getBytes());
                }
            }
        });
        server.start();
    }
}
