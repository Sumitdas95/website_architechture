# Template Maintenance

This project may not follow the best practices, implementations, or style that is used throughout Deliveroo. This does
not
mean this project is not a good starting point. Members of the team are currently looking at ways to improve this
to be in line with current Deliveroo standards and be the ideal Go service template going forward.

A work in progress Go in Deliveroo/Best Practices can be found
within [go/docs](https://deliveroo.atlassian.net/wiki/spaces/EN/pages/3856604593/Golang+Deliveroo).
# test-sonarqube

This template aims to provide a standard way to develop a new Go service at
Deliveroo.

It's intended to allow you to get up and running quickly, and to do so in a way
that means your service is consistent with other Go applications at Deliveroo.

## Quickstart guide

_Pre-requisites: [Go](https://golang.org/doc/install#install), [Homebrew](https://brew.sh/) and `roo-cli` (`brew install deliveroo/tools/roo-project`)._

Setting up a new service with the template can be done in just a few steps:

1. Create your new service:

    ```bash
    roo project new --lang=go --type=service name-of-service

    # Your project name-of-service was created at ~/go/src/github.com/deliveroo/name-of-service
    ```

2. Start the server for local development. This will install all development
dependencies and start the docker container of the Postgres database, so the
first time you run this it might take a minute or two:

    ```bash
    make start

    # >> starting...
    #  Server has booted! Listening on port 3000
    ```

3. To verify that the server is working correctly, query the example endpoint:

    ```bash
    curl --request GET --url http://localhost:3000/orders/1
    # {"ID":1,"Status":"NEW"}
    ```

4. Alternatively you can run the tests:

    ```bash
    make test
    ```

5. Now you can make changes to the project, and the server will automatically
restart whenever you make a change. You're all set up for local development!

6. Please help us improve the template! If you find bugs or missing features
please continue reading to find out how to contribute.

### Architectural considerations

There are several decisions that you may want to consider when starting from this blank slate. They don't all need to be resolved, but they hopefully will give some initial thoughts for how to start your project.

- If you're building a RESTful API, how will your API documentation be written?
  - Option: Using [OpenAPI](https://openapis.org). See more details at [go/openapi](http://go/openapi)
- Do you plan on using separate reader/writer nodes?
  - Option: Make sure that your application considers the `DATABASE_URL` and `DATABASE_URL_READER` environment variables injected by Hopper

## What is included?

The service template contains essentials such as:

* A CircleCI configuration that builds the project, runs the tests and runs a
linter.
* A Makefile to use for local development. Run `make help` to see all the
commands.
* Integration with [`apm-go`](https://github.com/deliveroo/apm-go) with support for
Datadog and Sentry.
* A simple Postgres setup, that works with CI and local development using Docker.
* Infrastructure configuration and control, more information within [docs/infrastructure](./docs/infrastructure.md).

And most importantly the service template has a standardized directory
structure:

* cmd/services -- this contains entry points to the service, e.g. for a `web` and a
`job-runner` binary.
* internal -- this includes all other code.
  * config -- code to configure the project using environment variables.
  * dependencies -- code to initialize the dependencies of the project.
  * orders -- an example of how to structure domain logic.
  * httpserver -- HTTP server logic, routes live here.
    * handlers -- REST endpoint handlers.

For more information about standard project layout at Deliveroo please see [go-project-structure](https://github.com/deliveroo/go-project-structure) repository.

 ## What is not yet included?

 There are many ways we could improve the template. One notable omission at the
 moment is that deploying your newly created service using CircleCI and Hopper
 is not yet automated. If you have suggestions on how to improve the project,
 please keep reading to find out how to get involved.

## Getting involved

There's lots to do and lots of decisions to make. We're trying to use the issue
tracker to help facilitate that discussion - please do get involved! You can
find us on Slack at
[#test-sonarqube](https://deliveroo.slack.com/messages/CHTQ08WAC).

Pull requests are open and welcome, but we'd encourage you to discuss ideas
first before putting in the time to implement them. It might be a good idea to
present some skeletons of how you'd implement a certain approach to illustrate
your ideas. However, it's best not to spend days crafting the perfect PR without
discussing it with others first.

## Guiding principles

We're guided by the following:

 - Most Go code at Deliveroo has yet to be written.
 - Different teams have been using Go in different ways. We want to take the
   best of all approaches.
 - We want to get something working, even if it's not perfect or we don't have
   complete consensus.
 - Just because something is on `main`, doesn't mean we're stuck with it. We
   can revisit decisions and consider alternatives.
 - We can evolve this over time.


## Should alternative approaches be considered?

Absolutely! Here, we're aiming to take what we think are the best of all the
existing approaches, and to mix into some improvements based on our collective
experiences. As set out in our [guiding principles](#guiding-principles), we're
aiming to get _something_ working. We can use that as a base to discuss
alternatives. We expect these discussions will work better if we're able to
compare how one approach works versus another.

This template is meant to guide you towards the "Deliveroo way", but not
constrain you. Bias towards consistency, but do the right thing to solve the
problem at hand.


## What if I don't need part of the template?

The template is (probably) going to contain everything you're going to need to
create a standard service at Deliveroo.

It might be that you won't need all of it for a service you're going to create.
That's fine! This is meant to be a boost to get started and to point you in the
right direction.

Eventually we hope to look into tooling which might allow you to generate a new
repository from this template using only the bits you need. We're prioritising
getting something to modularise later, but if you have experience with such
tools, please do raise an issue with your experiences!


## What about existing services?

We've already got a bunch of different services in Go at Deliveroo. We're not
proposing re-writing them to match this new service template - our time is
better spent adding new features for our riders/customers/restaurants!

However, if this becomes _the way_ that we design new services at Deliveroo, our
hope is that we can slowly make changes to our existing services to more closely
align with these standards, as we work on other feature work.

In any case, we believe that the majority of Go code at Deliveroo has yet to be
written. Creating a common foundation for the future is the most important thing
we're trying to address.


## Making a library?

You might also be interested in
[`lib-template-go`](https://github.com/deliveroo/lib-template-go), a template
for created a new shared library in Go!


## Prior Art

This isn't the first attempt at writing a template for Go applications. There's
already
[`go-service-template`](https://github.com/deliveroo/go-service-template), which
was used by a lot of apps within the Restaurant group.

There's a lot of really great things in that template, and we may choose to make
the same choices here. We're starting from scratch as it'll allow us to have
some discussion around each decision as we make it, and draw from our collective
experience writing Go apps in different teams.

## Using Go modules with private Deliveroo packages

The codebase depends on private Deliveroo packages such as determinator-go.
These have been "vendored" into the the `vendor` folder to, among other things,
avoid unnecessary network access (e.g. when running CI).

Whether you are maintaining this template code or any repositories
generated using this, when you need to update these to newer versions or you
need to depend on additional private packages, go will need to fetch these from
GitHub and vendor them too.

In order to fetch packages go will use git, and git must therefore know how
to authenticate with GitHub. By default HTTPS URLs are assumed, so go
modules operations will fail if you are using SSH to authenticate with GitHub.

There are two options:

* To continue to use HTTPS URLs, [create a personal access token](https://github.com/settings/tokens),
  make sure it is authorized ("enabled") for Deliveroo SSO.
  Copy the token and set it in an environment called `GITHUB_TOKEN` (assuming
  you are using Bash, put this in your `~/.bash_profile` if you want the
  variable to persist).
  Next run the following to configure git to use it for Deliveroo repositories:
  ```bash
  git config --global url."https://${GITHUB_TOKEN}:x-oauth-basic@github.com/deliveroo".insteadOf "https://github.com/deliveroo"
  ```
* To use SSH you need to configure git to rewrite any deliveroo HTTPS URLs
  to SSH format before git tries to fetch them:
  ```bash
  git config --global url."git@github.com:deliveroo".insteadOf "https://github.com/deliveroo"
    ```

You can test the configuration like this:

```bash
git ls-remote --get-url https://github.com/deliveroo/test-sonarqube.git
```

## How to register pgx codecs

The Go language does not have all the same data types as PostgreSQL. For example, Postgres has a `uuid` type but Go does not have a standard `uuid` type. There are 3rd party libraries available for these non-standard types, but `pgx` does not use them by default, to avoid external dependencies.

However, `pgx` supports both the ability to register new data types and to override the default handling of currently supported types.

In order to support the `uuid` type, for example, you can import the following package:

```go
import (
	pgxuuid "github.com/jackc/pgx-gofrs-uuid"
)
```

Then register the data type for your connection. If you are using a connection pool this should be done in the AfterConnect hook.

```go
dbconfig, err := pgxpool.ParseConfig(databaseURL)
if err != nil {
	// handle error
}
dbconfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
	pgxuuid.Register(conn.TypeMap())
	return nil
}
```