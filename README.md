# Baden Forecast Channel data

This repository generates a retail-compatible Wii Forecast Channel globe containing Nintendo's international location set plus the complete 21-place Limmattal set shown on the Regionale 2025 map. Brugg, Birmenstorf, and Dättwil are included as nearby additions. San Giovanni in Fiore, Maglie, Otranto, Santa Maria di Leuca, Gallipoli, Minervino di Lecce, Lecce, Bari, Brindisi, Crotone, and Cosenza remain included. Every custom Swiss and Italian location has maximum map display priority. GitHub Pages publishes signed forecast files for PAL language codes 1–6.

The Limmattal set is Untersiggenthal, Turgi, Gebenstorf, Baden, Obersiggenthal, Ennetbaden, Wettingen, Neuenhof, Killwangen, Würenlos, Spreitenbach, Oetwil an der Limmat, Geroldswil, Weiningen, Unterengstringen, Dietikon, Bergdietikon, Urdorf, Schlieren, Oberengstringen, and Zürich.

Weather data and most coordinates come from [Open-Meteo](https://open-meteo.com/). The geographic scope follows the official [Regionale 2025 map](https://regionale2025.ch/wp-content/uploads/2025/02/R2025_Kurzportrait_2022_online.pdf).

The generator is based on WiiLink's MPL-2.0-licensed [ForecastChannel](https://github.com/WiiLink24/ForecastChannel) project. It uses Open-Meteo instead of an AccuWeather key and gives each generated forecast a six-hour validity window. GitHub Actions refreshes and deploys the data every four hours; the overlap keeps the channel working if a scheduled run is delayed while preventing the Wii from retaining weather for half a day. An earlier 4,038-location experiment was removed because the retail channel downloaded the data but rejected it at runtime.

The matching WAD fixes the feed country code to `108`, so the complete custom location set and its weather icons load regardless of the country selected in Wii settings. It downloads the compact aliases below over plain HTTP, which the Wii's legacy networking supports:

```text
http://b7.185.199.108.153.nip.io/%d/108/f.bin
http://b7.185.199.108.153.nip.io/%d/108/s.bin
```

The repository and its Pages deployment must remain available for the custom channel to update.
