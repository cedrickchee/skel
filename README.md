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
