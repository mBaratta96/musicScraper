package metallum_cli;

import org.jsoup.Jsoup;
import org.jsoup.nodes.Document;
import java.io.IOException;

/**
 * Hello world!
 *
 */
public class Metallum {
    private static String metallum_url = "https://www.metal-archives.com/";
    private static String band_url = metallum_url.concat("bands/");

    public static void main(String[] args) {
        String test_url = band_url.concat("Panphage");
        try {
            Document metallum_page = Jsoup.connect(test_url).get();
            System.out.println(metallum_page.title());
        } catch (IOException e) {
            System.out.println(e);
            return;
        }
    }
}
