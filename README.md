# Tiny Assistant

Tiny Assistant is an LLM-based assistant that runs on your Pebble smartwatch,
if you still have a smartwatch that ceased production in 2016 lying around.

![A screenshot from a Pebble smartwatch running the Tiny Assistant. The user asked for the time, the assistant responded that it was 3:59 PM.](./docs/screenshot.png)

## Usage

### Server

To use Tiny Assistant, you will need to run the server in `service/` somewhere
your phone can reach.

You will need to set a few environment variables:

- `API_KEY` - an API key for your LLM provider
- `MODEL` - the model to use (e.g. `gpt-4o`, `claude-sonnet-4-20250514`)
- `BASE_URL` - (optional) the base URL of your OpenAI-compatible API endpoint.
  Defaults to OpenAI's API. Set this when using a BYOK server or a different
  provider that exposes an OpenAI-compatible API.
- `REDIS_URL` - a URL for a functioning Redis server. No data is persisted
  long-term, so a purely in-memory server is fine.
- `USER_IDENTIFICATION_URL` - a URL pointing to an instance of
  [user-identifier](https://github.com/pebble-dev/user-identifier).
- `MAPBOX_KEY` - an API key for [Mapbox](https://www.mapbox.com), which is
  used for geocoding. If no key is provided, geocoding will be unavailable.

### Client

Update the URL in `app/src/pkjs/urls.js` to point at your instance of the
server.

Then you can simply build it using the Pebble SDK and install on your watch.

## Contributing

See [`CONTRIBUTING.md`](CONTRIBUTING.md) for details.

## License

Apache 2.0; see [`LICENSE`](LICENSE) for details.

## Disclaimer

This project is not an official Google project. It is not supported by
Google and Google specifically disclaims all warranties as to its quality,
merchantability, or fitness for a particular purpose.
