# Baden Forecast Channel data

This repository generates a retail-compatible Wii Forecast Channel globe containing Baden and Nintendo's international location set. San Giovanni in Fiore, Maglie, Otranto, Santa Maria di Leuca, Gallipoli, Minervino di Lecce, Lecce, Bari, Brindisi, Crotone, and Cosenza are added at high display priority. GitHub Pages publishes signed forecast files for PAL language codes 1–6.

Weather data comes from [Open-Meteo](https://open-meteo.com/). Baden's coordinates come from [GeoNames](https://www.geonames.org/2661646/baden.html).

The generator is based on WiiLink's MPL-2.0-licensed [ForecastChannel](https://github.com/WiiLink24/ForecastChannel) project. It uses Open-Meteo instead of an AccuWeather key and gives each generated forecast a thirteen-hour validity window. GitHub Actions refreshes and deploys the data every four hours; the overlap keeps the channel working even if a scheduled run is delayed. An earlier 4,038-location experiment was removed because the retail channel downloaded the data but rejected it at runtime.

The matching WAD fixes the feed country code to `108`, so the complete custom location set and its weather icons load regardless of the country selected in Wii settings. It downloads the compact aliases below over plain HTTP, which the Wii's legacy networking supports:

```text
http://b7.185.199.108.153.nip.io/%d/108/f.bin
http://b7.185.199.108.153.nip.io/%d/108/s.bin
```

The repository and its Pages deployment must remain available for the custom channel to update.
