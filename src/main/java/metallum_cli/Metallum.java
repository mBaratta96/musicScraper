package metallum_cli;

import org.jsoup.Jsoup;
import org.jsoup.nodes.Document;
import org.jsoup.nodes.Element;
import org.jsoup.select.Elements;
import java.io.IOException;
import java.net.http.HttpRequest;

/**
 * Hello world!
 *
 */
public class Metallum {
    private static String metallum_url = "https://www.metal-archives.com/";
    private static String band_url = metallum_url.concat("bands/");

    private String getWebpage(String band) throws IOException {
        String test_url = band_url.concat(band);
        try {
            Document metallum_page = Jsoup.connect(test_url).get();
            Element discography = metallum_page.getElementById("band_disco");
            Element complete_discography = discography.getElementsByTag("a").select(":contains(Complete)").first();
            return complete_discography.attr("href");
        } catch (IOException e) {
            throw e;
        }
    }

    public static void main(String[] args) {
        Metallum app = new Metallum();
        try {
            String discography_url = app.getWebpage("Panphage");
            System.out.println(discography_url);
            Document discography_table = Jsoup.connect(discography_url).get();
            System.out.println(discography_table);
        } catch (IOException e) {
            System.err.println(e);
            return;
        }
    }
}
