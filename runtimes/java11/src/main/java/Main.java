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
        var client = HttpClient.newHttpClient();
        server.setExecutor(Executors.newSingleThreadExecutor());
        server.createContext("/messages", httpExchange -> {
            // read all input bytes
            var in = new InputStreamReader(httpExchange.getRequestBody(), StandardCharsets.UTF_8);
            var b = 0;
            var buf = new ByteArrayOutputStream();
            while ((b = in.read()) != -1) {
                buf.write((char) b);
            }
            in.close();
            try {
                for (byte[] msg : Handler.Handle(buf.toByteArray())) {
                    var resp = client.send(HttpRequest.newBuilder()
                            .uri(URI.create("http://localhost:3569/messages"))
                            .POST(HttpRequest.BodyPublishers.ofByteArray(msg))
                            .build(), HttpResponse.BodyHandlers.ofByteArray());
                    if (resp.statusCode() != 200) {
                        throw new Exception("non-200 error: " + resp.statusCode());
                    }
                }
                httpExchange.sendResponseHeaders(200, 0);
            } catch (Exception e) {
                System.err.println("failed to handle message: " + e.toString());
                e.printStackTrace();
                httpExchange.sendResponseHeaders(500, 0);
            }
        });
        server.start();
    }
}
