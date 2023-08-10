# Upfluence posts stats

This service provides an API for real-time analysis of social events (posts).

It is implemented with a Golang HTTP server.

## Usage

Requirements:

- [Go 1.21](https://golang.org/dl/)

Run the server:

```
go run github.com/xpetit/upfluence-social/cmd/server@latest
```

Try it:

```
curl -G localhost:8080/analysis -d dimension=likes -d duration=15s
```

Output:

> ```json
> {
>   "likes_p50": 18,
>   "likes_p90": 1181,
>   "likes_p99": 30753,
>   "maximum_timestamp": 1691656364,
>   "minimum_timestamp": 1570265971,
>   "total_posts": 39
> }
> ```

## Draft notes

I've let a few `TODO` in the code to discuss things in review.

### TODO

- Add tests
- Add benchmarks
- Add comments
- Improve documentation
- Structured logging (JSON Lines) with [`slog`](https://pkg.go.dev/log/slog) standard library.
  - Put the service name in each log output
  - Appropriate default level
  - Set `http.Server.ErrorLog` logger
  - Add debug logs for HTTP incoming requests
- Graceful shutdown?
- Handle duplicate events (same ID)?

## Implementation details & improvements

Make the event stream retryable with an exponential backoff.

A 32-bits unsigned integer was choosed for the counts because none of the measured dimensions exceeds 4 billion, but if one day "views" were added, this might not be enough (some YouTube videos have been watched more than 4 billion times).

To improve performance, consider using algorithms (t-digest?) that don't require to store & sort all the values to compute percentiles.

There are other optimization opportunities like using SOA (each dimension in a separate array) with delta-compression.

If it is acceptable to limit the analysis to the last 185 days, `uint16` can represent seconds (instead of `uint32`).
