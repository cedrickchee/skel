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

### Automated Version Numbers

We leverage Git to generate automated version numbers as part of your build
process. The process "burn-in" a version number and build time to your
application when building the binary.

The steps are: git commit recent changes, run `make build/api`, and then
check the version number in your binaries. Like so:

```sh
$ git add .
$ git commit -m 'generate version number automatically'

$ make build/api
Building cmd/api...
go build -ldflags="-s -X main.buildTime=2021-09-07T12:07:47+08:00 -X main.version=018442b" -o=./bin/api ./cmd/api
GOOS=linux GOARCH=amd64 go build -ldflags="-s -X main.buildTime=2021-09-07T12:07:47+08:00 -X main.version=018442b" -o=./bin/linux_amd64/api ./cmd/api

$ ./bin/api -version
Version:    018442b
Build time: 2021-09-07T12:07:47+08:00
```

We can see that our binary is now reporting that it was been built from a clean version of the repository with the commit hash `018442b`. Let’s cross check this against the `git log` output for the project:

```sh
$ git log
commit 018442b79eec04e2c739d5b3ab1ac4ca7345f17f (HEAD -> main)
Author: Cedric Chee <cedric@no-reply-github.com>
Date:   Tue Sep 7 12:07:01 2021 +0800

    generate version number automatically

...
```

The commit hash in our Git history aligns perfectly with our application version
number. And that means it’s now easy for us to identify exactly what code a
particular binary contains — all we need to do is run the binary with the
`-version` flag and then cross-reference it against the Git repository history.

## Deployment and Hosting

We're going to deploy our API application to a production server and expose it
on the Internet.

Every project and project team will have different technical and business needs
in terms of hosting and deployment, so it's impossible to lay out a
one-size-fits-all approach here.

We'll focus on hosting the application on a self-managed Linux server and using
standard Linux tooling to manage server configuration and deployment.

We'll also be automating the server configuration and deployment process as much
as possible, so that it's easy to make _continuous deployments_ and possible to
_replicate the server_ again in the future if you need to.

We'll be using [Digital Ocean](https://www.digitalocean.com/) as the hosting
provider.

In terms of infrastructure and architecture, we'll run everything on a single
Ubuntu Linux server. Our stack will consist of a PostgreSQL database and the
executable binary for our Skel API, operating in much the same way that we have
seen so far. But in addition to this, we'll also run
[Caddy](https://caddyserver.com/) as a _reverse proxy_ in front of the Skel API.

Using Caddy has a couple of benefits. It will automatically handle and terminate
HTTPS connections for us — including automatically generating and managing TLS
certificates via [Let's Encrypt](https://letsencrypt.org/) — and we can also use
Caddy to easily restrict internet access to our metrics endpoint.

You are going to:
- Provision an Ubuntu Linux server running on Digital Ocean to host your application.
- Automate the configuration of the server — including creating user accounts, configuring the firewall and installing necessary software.
- Automate the process of updating your application and deploying changes to the server.
- Run your application as a background service using [systemd](https://en.wikipedia.org/wiki/Systemd), as a non-root user.
- Use Caddy as a reverse proxy in front of your application to automatically manage TLS certificates and handle HTTPS connections.

### Server Configuration and Installing Software

Now that our Ubuntu Linux droplet has been successfully commissioned, we need to
do some housekeeping to secure the server and get it ready-to-use. Rather than
do this manually, we're going to create and use a reusable script to automate
these setup tasks.

You can find that regular Bash script in the `scripts/setup` folder in your
project directory. The script file is called [`01.sh`](./scripts/setup/01.sh).

**Prerequisite**
_This step ensure your Digital Ocean droplet is up and running and you've been able to successfully connect to it over SSH._

In order to log in to droplets in your Digital Ocean account you'll need a SSH keypair.

> **Suggestion:** If you're unfamiliar with SSH, SSH keys, or public-key cryptography generally, then I recommend reading through the first half of [this guide](https://www.digitalocean.com/community/tutorials/ssh-essentials-working-with-ssh-servers-clients-and-keys) to get an overview before continuing.

Open a new terminal window and try connecting to the droplet via SSH as the
`root` user, using the droplet IP address. Like so:

```sh
$ ssh root@{insert your VM IP}
The authenticity of host 'X.X.X.X (X.X.X.X)' can't be established.
  ...
Welcome to Ubuntu 20.04.1 LTS (GNU/Linux 5.4.0-51-generic x86_64)
  ...
root@skel-production:~# exit
logout
Connection to X.X.X.X closed.
```

OK, let's now run this script on our new Digital Ocean droplet. This will be a
two-step process:
1. First we need to copy the script to the droplet (which we will do using [rsync](https://linux.die.net/man/1/rsync)).
2. Then we need to connect to the droplet via SSH and execute the script.

Go ahead and run the following command to `rsync` the contents of the
`/scripts/setup` folder to the root user's home directory on the droplet.
Remember to replace the IP address with your own!

```sh
$ rsync -rP --delete ./scripts/setup root@X.X.X.X:/root
sending incremental file list
setup/
setup/01.sh
          3,327 100%    0.00kB/s    0:00:00 (xfr#1, to-chk=0/2)
```

Now that a copy of our setup script is on the droplet, let's use the `ssh`
command to execute the script on the remote machine as the `root` user.

Go ahead and run the script, entering a password for the `skel` _PostgreSQL
user_, like so:

```sh
$ ssh -t root@X.X.X.X 'bash /root/setup/01.sh'
Enter password for skel DB user: mySuprs3cretPASS0121
'universe' distribution component enabled for all sources.
Hit:1 http://security.ubuntu.com/ubuntu focal-security InRelease
Hit:2 https://repos.insights.digitalocean.com/apt/do-agent main InRelease
     ...
Script complete! Rebooting...
Connection to X.X.X.X closed by remote host.
Connection to X.X.X.X closed.
```

#### Connecting to the Droplet

After waiting a minute for the reboot to complete, try connecting to the droplet
as the `skel` user over SSH. This should work correctly (and the SSH key pair
you created previously should be used to authenticate the connection) but you
will be prompted to set a password.

To make connecting to the server a bit easier, and so we don't have to remember
the IP address, we add a Makefile rule for initializing a SSH connection to the
server as the `skel` user. Like so:

```
# Makefile
...

# ============================================================================ #
# PRODUCTION
# ============================================================================ #

production_host_ip = "INSERT YOUR IP ADDRESS"

## production/connect: connect to the production server
.PHONY: production/connect
production/connect:
	ssh skel@${production_host_ip}
```

_Remember to replace the IP address with your own._

You can then connect to your droplet whenever you need to by simply typing:

```sh
$ make production/connect
ssh skel@'X.X.X.X'
Welcome to Ubuntu 20.04.2 LTS (GNU/Linux 5.4.0-65-generic x86_64)
  ...
skel@skel-production:~$
```
