
# go-geofeed
Generate a [RFC8805](https://datatracker.ietf.org/doc/html/rfc8805) compliant geolocation feed using information from the [RIPE database](https://www.ripe.net).

This information should preferably be gathered from internal databases and not by using an external resource.

## Configuration
Configuration is done by setting environment variables.

Required:
* `NETWORKS`
	* Comma-separated list of networks to be queried. Example: 192.0.2.0/,2001:db8::/32
* `EMAIL`
	* Your e-mail address. It will be sent in the User-Agent-Field so the database operators has a way of contacting you

Optional:
* `LISTEN_ADDRESS` (default ":8080")
* `REFRESH_INTERVAL_MIN` (default 24)
* `REFRESH_INTERVAL_MAX` (default 36)
	* The geofeed will be regenerated on a random hour between these two values
* `KEY` (default empty)
	* Set to allow access to regenerate endpoint

## Usage
Set the environment variables and start the application. It will immediately start querying the database and generate the geofeed, during which it will return a `503` error to any client.
Once the feed is generated it will be regenerated either on-demand using the regenerate endpoint or on a random hour between the `REFRESH_INTERVAL_MIN` and `REFRESH_INTERVAL_MAX` values.

It's very much recommended to use a proxy (for instance nginx) in front to provide HTTPS capabilities.

### Endpoints
* `/geofeed.csv` returns the geofeed
* `/regenerate` will regenerate the geofeed immediately if the `KEY` environment variable is set and the `X-Geofeed-Key` header is sent
* Any other URL gives a `301 Moved Permanently` to `/geofeed.csv`