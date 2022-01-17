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

## Prerequisites

You'll need to install these softwares and tools on your machine:

- [Go 1.16 or newer](https://golang.org/dl/)
- PostgreSQL database (version 12 or newer)
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
Usage of ./bin/linux_amd64/api:
  -cors-trusted-origins value
    	Trusted CORS origins (space separated)
  -db-dsn string
    	PostgreSQL DSN
  -db-max-idle-conns int
    	PostgreSQL max idle connections (default 25)
  -db-max-idle-time string
    	PostgreSQL max idle time (default "15m")
  -db-max-open-conns int
    	PostgreSQL max open connections (default 25)
  -env string
    	Environment (development|staging|production) (default "development")
  -limiter-burst int
    	Rate limiter maximum burst (default 4)
  -limiter-enabled
    	Enable rate limiter (default true)
  -limiter-rps float
    	Rate limiter maximum requests per second (default 2)
  -port int
    	API server port (default 4000)
  -smtp-host string
    	SMTP host (default "smtp.mailtrap.io")
  -smtp-password string
    	SMTP password (default "xxxxxxxxxxxxxx")
  -smtp-port int
    	SMTP port (default 2525)
  -smtp-sender string
    	SMTP sender (default "Skel <no-reply@example.com>")
  -smtp-username string
    	SMTP username (default "xxxxxxxxxxxxxx")
  -version
    	Display version and exit
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
  help                               print this help message
  run/api                            run the cmd/api application
  db/psql                            connect to the database using psql
  db/migrations/new name=$1          create a new database migration
  db/migrations/up                   apply all up database migrations
  audit                              tidy and vendor dependencies and format, vet and test all code
  vendor                             tidy and vendor dependencies
  build/api                          build the cmd/api application
  production/connect                 connect to the production server
  production/deploy/api              deploy the api to production
  production/configure/api.service   configure the production systemd api.service file
  production/configure/caddyfile     configure the production Caddyfile
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

### Profiling Test Coverage

A feature of the `go test` tool is the metrics and visualizations that it
provides for test coverage.

```sh
$ make coverage
Running test coverage ...
go test -cover ./...
ok      github.com/cedrickchee/skel/cmd/api     (cached)        coverage: 18.9% of statements
?       github.com/cedrickchee/skel/cmd/examples/cors/preflight [no test files]
?       github.com/cedrickchee/skel/cmd/examples/cors/simple    [no test files]
ok      github.com/cedrickchee/skel/internal/data       (cached)        coverage: 3.3% of statements
?       github.com/cedrickchee/skel/internal/jsonlog    [no test files]
?       github.com/cedrickchee/skel/internal/mailer     [no test files]
?       github.com/cedrickchee/skel/internal/validator  [no test files]
go test -covermode=count -coverprofile=/tmp/profile.out ./...
ok      github.com/cedrickchee/skel/cmd/api     0.008s  coverage: 18.9% of statements
?       github.com/cedrickchee/skel/cmd/examples/cors/preflight [no test files]
?       github.com/cedrickchee/skel/cmd/examples/cors/simple    [no test files]
ok      github.com/cedrickchee/skel/internal/data       0.115s  coverage: 3.3% of statements
?       github.com/cedrickchee/skel/internal/jsonlog    [no test files]
?       github.com/cedrickchee/skel/internal/mailer     [no test files]
?       github.com/cedrickchee/skel/internal/validator  [no test files]
go tool cover -html=/tmp/profile.out
```

From the results here we can see that 18.9% of the statements in our `cmd/api`
package are executed during our tests, and for our `internal/data` the figure is
3.3%.

This will open a browser window containing a navigable and highlighted
representation of your code.

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

We can see that our binary is now reporting that it was been built from a clean version of the repository with the commit hash `018442b`. Let's cross check this against the `git log` output for the project:

```sh
$ git log
commit 018442b79eec04e2c739d5b3ab1ac4ca7345f17f (HEAD -> main)
Author: Cedric Chee <cedric@no-reply-github.com>
Date:   Tue Sep 7 12:07:01 2021 +0800

    generate version number automatically

...
```

The commit hash in our Git history aligns perfectly with our application version
number. And that means it's now easy for us to identify exactly what code a
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

### Deployment and Running Application

At this point our droplet is set up with all the software and user accounts that we need, so let's move on to the process of deploying and running our API application.

At a very high-level, our deployment process will consist of three actions:

1. Copying the application binary and SQL migration files to the droplet.
2. Executing the migrations against the PostgreSQL database on the droplet.
3. Starting the application binary as a _background service_.

To execute the first two steps automatically, we made a `production/deploy/api`
rule in Makefile.

```sh
$ make production/deploy/api
rsync -rP --delete ./bin/linux_amd64/api ./migrations skel@"X.X.X.X":~
sending incremental file list
api
      7,618,560 100%  119.34kB/s    0:01:02 (xfr#1, to-chk=13/14)
migrations/
migrations/000001_create_movies_table.down.sql
             28 100%   27.34kB/s    0:00:00 (xfr#2, to-chk=11/14)
migrations/000001_create_movies_table.up.sql
            286 100%  279.30kB/s    0:00:00 (xfr#3, to-chk=10/14)
migrations/000002_add_movies_check_constraints.down.sql
            198 100%  193.36kB/s    0:00:00 (xfr#4, to-chk=9/14)
migrations/000002_add_movies_check_constraints.up.sql
            289 100%  282.23kB/s    0:00:00 (xfr#5, to-chk=8/14)
migrations/000003_add_movies_indexes.down.sql
             78 100%   76.17kB/s    0:00:00 (xfr#6, to-chk=7/14)
migrations/000003_add_movies_indexes.up.sql
            170 100%  166.02kB/s    0:00:00 (xfr#7, to-chk=6/14)
migrations/000004_create_users_table.down.sql
             27 100%   26.37kB/s    0:00:00 (xfr#8, to-chk=5/14)
migrations/000004_create_users_table.up.sql
            294 100%  287.11kB/s    0:00:00 (xfr#9, to-chk=4/14)
migrations/000005_create_tokens_table.down.sql
             28 100%   27.34kB/s    0:00:00 (xfr#10, to-chk=3/14)
migrations/000005_create_tokens_table.up.sql
            203 100%   99.12kB/s    0:00:00 (xfr#11, to-chk=2/14)
migrations/000006_add_permissions.down.sql
             73 100%   35.64kB/s    0:00:00 (xfr#12, to-chk=1/14)
migrations/000006_add_permissions.up.sql
            452 100%  220.70kB/s    0:00:00 (xfr#13, to-chk=0/14)
ssh -t skel@"X.X.X.X" "migrate -path ~/migrations -database $SKEL_DB_DSN up"
1/u create_movies_table (11.782733ms)
2/u add_movies_check_constraints (23.109006ms)
3/u add_movies_indexes (30.61223ms)
4/u create_users_table (39.890662ms)
5/u create_tokens_table (48.659641ms)
6/u add_permissions (58.23243ms)
Connection to X.X.X.X closed.
```

#### Running the API as a Background Service

The next step is to configure it to run as a _background service_, including
starting up automatically when the droplet is rebooted.

We do this using [systemd](https://www.freedesktop.org/wiki/Software/systemd/).

We've made a unit file
([`scripts/production/api.service`](./scripts/production/api.service)), which
informs systemd how and when to run the service.

The next step is to install this unit file on our droplet and start up the
service.

Go ahead and run the Makefile rule. The output you see should look similar to
this:

```sh
$ make production/configure/api.service
rsync -P ./scripts/production/api.service skel@"X.X.X.X":~
sending incremental file list
api.service
          1,266 100%    0.00kB/s    0:00:00 (xfr#1, to-chk=0/1)
ssh -t skel@"X.X.X.X" '\
        sudo mv ~/api.service /etc/systemd/system/ \
        && sudo systemctl enable api \
        && sudo systemctl restart api \
'
[sudo] password for skel:
Created symlink /etc/systemd/system/multi-user.target.wants/api.service → /etc/systemd/system/api.service.
Connection to X.X.X.X closed.
```

Next connect to the droplet and check the status of the new `api` service using
the `sudo systemctl status api` command:

```sh
$ make production/connect
skel@skel-production:~$ sudo systemctl status api
● api.service - Skel API service
     Loaded: loaded (/etc/systemd/system/api.service; enabled; vendor preset: enabled)
     Active: active (running) since Mon 2021-09-09 03:15:35 SGT; 1min 31s ago
   Main PID: 6891 (api)
      Tasks: 6 (limit: 1136)
     Memory: 1.8M
     CGroup: /system.slice/api.service
             └─6997 /home/skel/api -port=4000 -db-dsn=postgres://skel:blahP@55blah@localhost/skel -env=production

Apr 09 03:15:35 skel-production systemd[1]: Starting Skel API service...
Apr 09 03:15:35 skel-production systemd[1]: Started Skel API service.
Apr 09 03:15:35 skel-production api[6891]: {"level":"INFO","time":"2021-09-09T07:15:35Z", ...}
Apr 09 03:15:35 skel-production api[6891]: {"level":"INFO","time":"2021-09-09T07:15:35Z", ...}
```

This confirms that the service is running successfully in the background and, in
my case, that it has the PID (process ID) `6891`.

#### Use Caddy as a Reverse Proxy

We're now in the state where we have Caddy running as a background service and
listening for HTTP requests on port `80`.

So the next step in setting up our production environment is to configure Caddy
to act as a reverse proxy and forward any HTTP requests that it receives onward
to our API.

To configure Caddy, we created a [Caddyfile](./scripts/production/Caddyfile).

If you're following along, please go ahead and replace the IP address in the
Caddyfile with the address of your own droplet (server).

Next deploy this Caddyfile into your droplet:

```sh
$ make production/configure/caddyfile
rsync -P ./scripts/production/Caddyfile skel@"X.X.X.X":~
sending incremental file list
Caddyfile
             53 100%    0.00kB/s    0:00:00 (xfr#1, to-chk=0/1)
ssh -t skel@"X.X.X.X" '\
        sudo mv ~/Caddyfile /etc/caddy/ \
        && sudo systemctl reload caddy \
'
[sudo] password for skel:
Connection to X.X.X.X closed.
```

You should see that the Caddyfile is copied across and the reload executes
cleanly without any errors.

At this point you can visit `http://<your_droplet_ip>/v1/healthcheck` in a web
browser, and you should find that the request is successfully forwarded on from
Caddy to our API.

## Application Metrics

The metrics are no longer publicly accessible, you can still access them by
connecting to your droplet via SSH.

You can open a SSH tunnel to the droplet and view them using a web browser on
your local machine. For example, you could open an SSH tunnel between port
`4000` on the droplet and port `9999` on your local machine by running the
following command (make sure to replace _both_ IP addresses with your own
droplet IP):

```sh
$ ssh -L :9999:X.X.X.X:4000 skel@X.X.X.X
```

While that tunnel is active, you should be able to visit
`http://localhost:9999/debug/vars` in your web browser and see your application
metrics.

## Using a Domain Name (Optional)

For the next step of our deployment, if you want, you can configure Caddy so
that you can access our droplet via a domain name, instead of needing to use the
IP address.

I'm going to use the domain `skel.cedricchee.com` in the sample code here, but
you should swap this out for your own domain if you're following along.

The first thing you'll need to do is configure the DNS records for your domain
name so that they contain an `A` record pointing to the IP address for your
droplet. So in my case the DNS record would look like this:

```
A     skel.cedricchee.com     X.X.X.X
```

> **Note:** If you're not sure how to alter your DNS records, your domain name registrar should provide guidance and documentation.

Once you've got the DNS record in place, the next task is to update the
Caddyfile to use your domain name instead of your droplet's IP address. Go ahead
and swap this out like so (remember to replace `skel.cedricchee.com` with your
own domain name):

```
http://skel.cedricchee.com {
    respond /debug/* "Not Permitted" 403
    reverse_proxy localhost:4000
}
```

And then redeploy the Caddyfile to your droplet again:

```sh
$ make production/configure/caddyfile
```

Once you've done that, you should now be able to access the API via your domain name by visiting `http://<your_domain_name>/v1/healthcheck` in your browser.

## Enabling HTTPS (Optional)

Now that we have a domain name set up we can utilize one of Caddy's headline
features: _automatic HTTPS_.

Caddy will automatically handle provisioning and renewing TLS certificates for
your domain via Let's Encrypt, as well as redirecting all HTTP requests to
HTTPS. It's simple to set up, very robust, and saves you the overhead of needing
to keep track of certificate renewals manually.

To enable this, we just need to update our `Caddyfile` so that it looks like
this:

```
# Set the email address that should be used to contact you if there is a
# problem with your TLS certificates.
{
  email you@example.com
}

# Remove the http:// prefix from your site address.
skel.cedricchee.com {
    respond /debug/* "Not Permitted" 403
    reverse_proxy localhost:4000
}
```

For the final time, deploy this Caddyfile update to your droplet.

```sh
$ make production/configure/caddyfile
```

And then when you refresh the page in your web browser, you should find that it
is automatically redirected to a HTTPS version of the page.

## Working with Branches

You can switch the authentication process for our API to use JSON Web Tokens
(JWTs). Just `git checkout` the feature branch named
["feat/jwt-auth"](https://github.com/cedrickchee/skel/tree/feat/jwt-auth).

> **Important:** Using JWTs for this particular application doesn't have any
> benefits over our current "stateful token" approach. The JWT approach is more
> complex, ultimately requires the same number of database lookups, and we lose
> the ability to revoke tokens. For those reasons, it doesn't make much sense to
> use them here. But I still want to explain the pattern for two reasons:
>
> - JWTs have a lot of mind-share, and as a developer it's likely that you will
>   run into existing codebases that use them for authentication, or that
>   outside influences mean you are forced to use them.
>
> - It's worth understanding how they work in case you need to implement APIs
>   that require _delegated authentication_, which _is_ a scenario that they're
>   useful in.

## License

<details>

<summary><b>Expand License</b></summary>

This repository contains a variety of content; some developed by Cedric Chee, 
and some from third-parties. The third-party content is distributed under the 
license provided by those parties.

*I am providing code and resources in this repository to you under an open 
source license. Because this is my personal repository, the license you receive 
to my code and resources is from me and not my employer.*

The content developed by Cedric Chee is distributed under the following license:

### Text

The text content is released under the CC-BY-NC-ND license.
Read more at [Creative Commons](https://creativecommons.org/licenses/by-nc-nd/3.0/us/legalcode).

### Code

The code in this repository is released under the [MIT license](LICENSE).
</details>
