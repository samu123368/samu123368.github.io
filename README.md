# Baden Forecast Channel data

This repository generates a retail-compatible Wii Forecast Channel globe containing Nintendo's international location set plus a compact Swiss set: Untersiggenthal, Baden, Brugg, Birmenstorf, Gebenstorf, Dättwil, Turgi, Wettingen, Spreitenbach, Dietikon, Schlieren, and Zürich. San Giovanni in Fiore, Maglie, Otranto, Santa Maria di Leuca, Gallipoli, Minervino di Lecce, Lecce, Bari, Brindisi, Crotone, and Cosenza remain included. Every custom Swiss and Italian location has maximum map display priority. GitHub Pages publishes signed forecast files for PAL language codes 1–6.

Weather data and coordinates come from [Open-Meteo](https://open-meteo.com/).

The generator is based on WiiLink's MPL-2.0-licensed [ForecastChannel](https://github.com/WiiLink24/ForecastChannel) project. It uses Open-Meteo instead of an AccuWeather key. GitHub Actions refreshes and deploys the data every hour, and each generated file expires after one hour so the channel requests the next hourly forecast. An earlier 4,038-location experiment was removed because the retail channel downloaded the data but rejected it at runtime.

The matching WAD fixes the feed country code to `108`, so the complete custom location set and its weather icons load regardless of the country selected in Wii settings. It downloads the compact aliases below over plain HTTP, which the Wii's legacy networking supports:

```text
http://b7.185.199.108.153.nip.io/%d/108/f.bin
http://b7.185.199.108.153.nip.io/%d/108/s.bin
```

The repository and its Pages deployment must remain available for the custom channel to update.
