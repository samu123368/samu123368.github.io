# Baden Forecast Channel data

This repository generates Wii Forecast Channel data for Switzerland with Baden, Aargau included. GitHub Pages publishes the generated `forecast.bin` and `short.bin` files for PAL language codes 1–6.

Weather data comes from [Open-Meteo](https://open-meteo.com/). Baden's coordinates come from [GeoNames](https://www.geonames.org/2661646/baden.html).

The generator is based on WiiLink's MPL-2.0-licensed [ForecastChannel](https://github.com/WiiLink24/ForecastChannel) project. It is restricted to Switzerland, uses Open-Meteo instead of an AccuWeather key, and gives each generated forecast a three-hour validity window. GitHub Actions refreshes and deploys the data every two hours.

The matching WAD is configured to download from:

```text
http://samu123368.github.io/%d/%03d/forecast.bin
http://samu123368.github.io/%d/%03d/short.bin
```

The repository and its Pages deployment must remain available for the custom channel to update.
