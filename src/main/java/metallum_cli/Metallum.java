package metallum_cli;

import java.util.List;
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
    private static final String METALLUM_URL = "https://www.metal-archives.com/";
    private static final String BAND_URL = METALLUM_URL.concat("bands/");

    private String getWebpage(String band) throws IOException {
        String test_url = BAND_URL.concat(band);
        Document metallum_page = Jsoup.connect(test_url).get();
        Element discography = metallum_page.getElementById("band_disco");
        Element complete_discography = discography.getElementsByTag("a").select(":contains(Complete)").first();
        return complete_discography.attr("href");
    }

    public static void main(String[] args) {
        Metallum app = new Metallum();
        try {
            String discography_url = app.getWebpage("Panphage");
            Document discography_table = Jsoup.connect(discography_url).get();
            Elements table_rows = discography_table.getElementsByTag("tr");

            table_rows.forEach(row -> {
                Elements row_values = (row.getElementsByTag("td").size() > 0) ? row.getElementsByTag("td")
                        : row.getElementsByTag("th");
                List<String> table_values = row_values.stream().map(value -> value.text()).toList();
                System.out.format("%32s%32s%32s%32s\n", table_values.get(0), table_values.get(1),
                        table_values.get(2),
                        table_values.get(3));
            });
        } catch (IOException e) {
            System.err.println(e);
            return;
        }
    }
}
