package metallum_cli;

import com.github.loki.afro.metallum.search.query.entity.BandQuery;
import com.github.loki.afro.metallum.search.API;;

/**
 * Hello world!
 *
 */
public class App {
    public static void main(String[] args) {
        BandQuery query = BandQuery.byName("slayer", false);
        API.getBandsFully(query).forEach(band -> {
            System.out.println("Bandname: " + band.getName());
            System.out.println("Bandgenre: " + band.getGenre());
            System.out.println("Bandstatus: " + band.getStatus());
            System.out.println("Partial Discs: " + band.getDiscsPartial());
            System.out.println("Discs: " + band.getDiscs());
            System.out.println("---");
        });
    }
}
