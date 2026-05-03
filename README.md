# Bobby Assistant

Bobby Assistant is an LLM-based assistant that runs on your Pebble smartwatch.

This is a fork of [Bobby Assistant](https://github.com/pebble-dev/bobby-assistant) with the following changes:

- **Self-hosted, no external dependencies.** Removed all reliance on Rebble.io services (user authentication, quota tracking, timeline API, feedback reporting). Users bring their own LLM API key and configure it directly on the watch.
- **OpenAI-compatible API.** Replaced the Google Gemini client with the OpenAI Go SDK, supporting any OpenAI-compatible endpoint (OpenAI, Ollama, LiteLLM, etc.).
- **SQLite instead of Redis.** Thread persistence and currency rate caching use a local SQLite database. No external services required.
- **Per-user LLM configuration.** API key, base URL, and model are configured in the watch app's settings and sent with each request. No server-side secrets needed for LLM access.
- **No quota system.** Removed credit-based usage tracking and per-user quotas.
- **Sentry integration.** Replaced Honeycomb tracing with Sentry for error reporting and performance monitoring.
- **Local reminders.** Reminders are stored on-device. Timeline pin notifications are no longer supported since that required Rebble authentication.

![A screenshot from a Pebble smartwatch running the assistant. The user asked for the time, the assistant responded that it was 3:59 PM.](./docs/screenshot.png)

## Usage

### Server

Run the server in `service/` somewhere your phone can reach.

Environment variables:

- `SENTRY_DSN` - (optional) Sentry DSN for error reporting
- `DB_PATH` - (optional) path to the SQLite database file. Defaults to `bobby.db`
- `MAPBOX_KEY` - (optional) API key for [Mapbox](https://www.mapbox.com), used for geocoding and POI search
- `IBM_KEY` - (optional) API key for IBM weather data
- `EXCHANGE_RATE_API_KEY` - (optional) API key for exchange rate lookups
- `GOOGLE_MAPS_STATIC_KEY`, `GOOGLE_MAPS_STATIC_SECRET`, `GOOGLE_MAPS_STATIC_MAP_ID` - (optional) for static map images

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

This project is not an official Google project. It is not supported by
Google and Google specifically disclaims all warranties as to its quality,
merchantability, or fitness for a particular purpose.
