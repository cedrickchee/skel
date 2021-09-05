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
