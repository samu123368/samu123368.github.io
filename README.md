# Baden Forecast Channel data

This repository generates Wii Forecast Channel data for Switzerland with Baden, Aargau included. GitHub Pages publishes signed forecast files for PAL language codes 1–6.

Weather data comes from [Open-Meteo](https://open-meteo.com/). Baden's coordinates come from [GeoNames](https://www.geonames.org/2661646/baden.html).

The generator is based on WiiLink's MPL-2.0-licensed [ForecastChannel](https://github.com/WiiLink24/ForecastChannel) project. It is restricted to Switzerland, uses Open-Meteo instead of an AccuWeather key, and gives each generated forecast a three-hour validity window. GitHub Actions refreshes and deploys the data every two hours.

The matching WAD is configured to download the compact aliases below over plain HTTP, which the Wii's legacy networking supports:

```text
http://b7.185.199.108.153.nip.io/%d/%03d/f.bin
http://b7.185.199.108.153.nip.io/%d/%03d/s.bin
```

The repository and its Pages deployment must remain available for the custom channel to update.
