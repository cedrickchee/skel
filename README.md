# Skel

Skel is an example application written in Go.

This repository can be used as a reference--the code is well documented.

It's a good fit as starter code for developing JSON APIs which act as backends
for Single-Page Applications (SPAs), mobile applications, or function as
stand-alone services.

The project addresses the design choices of developing applications:

- Project structure and organization.
- Practical code patterns for creating robust and maintainable programs.

Development practices is based on guiding principles of well written Go code.
- Clarity
- Simplicity
- Correctness
- Productivity

## Development

You can build `skel` locally by cloning the repository, then run:

```sh
$ make
```

### Development Environment with Live Code Reloading

Run [reflex](https://github.com/cespare/reflex) which will be used for hot
recompiling the code. It can be very useful for quickly testing your changes.

```sh
$ reflex -c reflex.conf
```

What the command means is: "watch for changes to `go.mod` and all files ending
in `.go` and execute `go run ./cmd/api` when it happens. The `-s` flag stands
for service and will make reflex kill previously run command before starting it
again, which is exactly what we want.

## Environment Variables

PostgreSQL database DSN.

```sh
$ export SKEL_DB_DSN=postgres://skel:pa55word@localhost/skel
```

## Command-Line Flags

```sh
$ go run ./cmd/api --help
Usage of /tmp/go-build2584491206/b001/exe/api:
  -db-dsn string
    	PostgreSQL DSN (default "postgres://skel:pa55word@localhost/skel")
  -db-max-idle-conns int
    	PostgreSQL max idle connections (default 25)
  -db-max-idle-time string
    	PostgreSQL max idle time (default "15m")
  -db-max-open-conns int
    	PostgreSQL max open connections (default 25)
  -env string
    	Environment (development|staging|production) (default "development")
  -port int
    	API server port (default 4000)
```
