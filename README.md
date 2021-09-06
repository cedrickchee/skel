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

## Prerequisite

You'll need to install these softwares and tools on your machine:

- PostgreSQL database (version 12+)
- make utility
- [reflex](https://github.com/cespare/reflex) (optional)
- [staticcheck](https://staticcheck.io/) tool to carry out some [additional static analysis checks](https://staticcheck.io/docs/checks)

## Quickstart

First, clone the repository:

```sh
$ git clone https://github.com/cedrickchee/skel.git
$ cd skel
```

**Environment Variables**

Create a `.envrc` file in the root directory of this project
by renaming `.envrc.example` to `.envrc`.

```sh
# PostgreSQL database DSN.
SKEL_DB_DSN=postgres://skel:pa55word@localhost/skel
```

Then, you can build and run `skel` by using `make`:

```sh
$ make run/api
```

## Development

### Development Environment with Live Code Reloading

Run reflex which will be used for hot recompiling the code. It can be very
useful for quickly testing your changes.

```sh
$ reflex -c reflex.conf
```

What the command means is: "watch for changes to `go.mod` and all files ending
in `.go` and execute `go run ./cmd/api` when it happens. The `-s` flag stands
for service and will make reflex kill previously run command before starting it
again, which is exactly what we want.

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

Start the API, passing in a couple of command line flags for different purposes:

- Run the API with rate limiting disabled:

  ```sh
  $ go run ./cmd/api -limiter-enabled=false
  ```

- Before you run your load testing, you might like to play around with this and
  try changing some of the configuration parameters for the connection pool to
  see how it affects the behavior of the figures (from your benchmarking tool)
  under load.

  ```sh
  $ go run ./cmd/api -limiter-enabled=false -db-max-open-conns=50 -db-max-idle-conns=50 -db-max-idle-time=20s -port=4000
  ```

- Run your API, passing in `http://localhost:9000` and `http://localhost:9001`
  as CORS trusted origins like so:

  ```sh
  $ go run ./cmd/api -cors-trusted-origins='http://localhost:9000 http://localhost:9001'
  ```

## Using Makefile

Use the GNU [make](https://www.gnu.org/software/make/manual/make.html) utility
and makefiles to help automate common tasks in out project, such as creating and
executing database migrations.

### Displaying help information

Execute the `help` target, you should get a response which lists all the
available targets and the corresponding help text.

```sh
$ make
Usage:
  help                        print this help message
  run/api                     run the cmd/api application
  db/psql                     connect to the database using psql
  db/migrations/new name=$1   create a new database migration
  db/migrations/up            apply all up database migrations
  audit                       tidy and vendor dependencies and format, vet and test all code
  vendor                      tidy and vendor dependencies
  build/api                   build the cmd/api application
```

### Using make for Common Tasks

You should be able to execute the rules by typing the full target name when
running `make`. For example:

```sh
$ make run/api
go run ./cmd/api
{"level":"INFO","time":"2021-09-04T14:44:36Z","message":"database connection pool established"}
{"level":"INFO","time":"2021-09-04T14:44:36Z","message":"starting server","properties":{"addr":":4000","env":"development"}}
```

If you run the `db/migrations/up` rule with the `name=create_example_table`
argument you should see the following output:

```sh
$ make db/migrations/up name=create_example_table
Creating migration files for create_example_table
migrate create -seq -ext=.sql -dir=./migrations create_example_table
/home/cedric/dev/repo/gh/skel/migrations/000007_create_example_table.up.sql
/home/cedric/dev/repo/gh/skel/migrations/000007_create_example_table.down.sql
```

## Quality Controlling Code

The audit rule will:
- prune dependencies
- verify module dependencies
- format all `.go` files, according to the Go standard
- vet code; runs a variety of analyzers which carry out static analysis
- staticcheck; carry out some additional static analysis checks
- test

To run these checks before you commit any code changes into your version control
system or build any binaries.

```sh
$ make audit
```

## Vendoring New Dependencies

**Note:** It's important to point out that there's no easy way to verify that
the _checksums of the vendored dependencies_ match the checksums in the `go.sum`
file.

To mitigate that, it's a good idea to run both `go mod verify` and
`go mod vendor` regularly.
Using `go mod verify` will verify that the dependencies in your module cache
match the `go.sum` file, and `go mod vendor` will copy those same dependencies
from the module cache into your `vendor` directory.

```sh
# vendor rule execute both `go mod verify` and `go mod vendor` commands.
$ make vendor
```

## Build

Build and run executable binaries for our applications.

Go supports **cross-compilation**, so you can generate a binary suitable for use
on a different machine.

Let's create two binaries — one for use on your local machine, and another for
deploying to the Ubuntu Linux server.

To build binaries, you need to execute:

```sh
$ make build/api
```

You should see that two binary files are now created — one for local machine at
`./bin/api`; with the cross-compiled binary located under the
`./bin/linux_amd64` directory.

And you should be able to run this executable to start your API application,
passing in any command-line flag values as necessary. For example:

```sh
$ ./bin/api -port=3000 -db-dsn=postgres://myuser:mysuperpass@localhost/skel
{"level":"INFO","time":"2021-09-06T13:00:00Z","message":"database connection pool established"}
{"level":"INFO","time":"2021-09-06T13:00:00Z","message":"starting server","properties":{"addr":":3000","env":"development"}}
```
