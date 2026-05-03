# Bobby Assistant

Bobby Assistant is an LLM-based assistant that runs on your Pebble smartwatch.

This is a fork of [Bobby Assistant](https://github.com/pebble-dev/bobby-assistant) with the following changes:

- **Self-hosted, no external account required.** Removed all reliance on Rebble.io services (user authentication, quota tracking, timeline API, feedback reporting). Users bring their own LLM API key and configure it directly on the watch.
- **OpenAI-compatible API.** Replaced the Google Gemini client with the OpenAI Go SDK, supporting any OpenAI-compatible endpoint (OpenAI, Ollama, LiteLLM, etc.).
- **SQLite instead of Redis.** Thread persistence uses a local SQLite database. No external services required.
- **Per-user LLM configuration.** API key, base URL, and model are configured in the watch app's settings and sent with each request. No server-side secrets needed for LLM access.
- **Consolidated Google Maps integration.** Replaced Mapbox (geocoding) and IBM Weather with Google's Geocoding and Weather APIs, using a single API key.
- **Sentry integration.** Replaced Honeycomb tracing with Sentry for error reporting and performance monitoring.
- **No quota system.** Removed credit-based usage tracking and per-user quotas.
- **No currency conversion.** Removed the exchange rate feature and its external API dependency.
- **Local reminders.** Reminders are stored on-device. Timeline pin notifications are no longer supported since that required Rebble authentication.

![A screenshot from a Pebble smartwatch running the assistant. The user asked for the time, the assistant responded that it was 3:59 PM.](./docs/screenshot.png)

## Usage

### Server

Run the server in `service/` somewhere your phone can reach.

Environment variables:

- `SENTRY_DSN` - (optional) Sentry DSN for error reporting
- `DB_PATH` - (optional) path to the SQLite database file. Defaults to `bobby.db`
- `GOOGLE_MAPS_STATIC_KEY`, `GOOGLE_MAPS_STATIC_SECRET`, `GOOGLE_MAPS_STATIC_MAP_ID` - (optional) Google Maps API key for geocoding, routing, POI search, weather, and static map images

### Client

Configure the following in the watch app's settings page:

- **Server URL** - the WebSocket URL of your server (e.g. `wss://my-server.com/query`)
- **API Key** - your OpenAI-compatible API key
- **Base URL** - (optional) base URL if using a non-OpenAI provider
- **Model** - the model to use (e.g. `gpt-4o-mini`)

Build with the Pebble SDK and install on your watch.

## Contributing

See [`CONTRIBUTING.md`](CONTRIBUTING.md) for details.

## License

Apache 2.0; see [`LICENSE`](LICENSE) for details.

## Disclaimer

This project is a fork of [Bobby Assistant](https://github.com/pebble-dev/bobby-assistant) by Google.
It is not affiliated with or endorsed by Google or Rebble.
