# nate-mlb
A web server which compares MLB baseball scores and NFL football scores.

Built with the [Go](https://github.com/golang/go) programming language.

Runs on a [Postgresql](https://github.com/postgres/postgres) database.


# Dependencies
* [pq](https://github.com/lib/pq) (Postgresql Driver)
* [Bootstrap](https://github.com/twbs/bootstrap) (css, html widgets)
* [jQuery](https://github.com/jquery/jquery) (Javascript functions for Bootstrap modals)
* [Font-Awesome](https://github.com/FortAwesome/Font-Awesome) (icons on about page)
* [bcrypt](https://github.com/golang/crypto) (password encryption)
* [dep](https://github.com/golang/dep) (Go dependency management tool) Run `dep ensure` to update dependencies.


# Installation

### Database

* **Postgresql** See [init.sql](sql/init.sql) for setup instructions on how to create the dabase, table creation code, and demo data.

* **Admin Password** See [init.sql](sql/init.sql) for how to set the admin password for the application.

### Set environment variables

* **PORT** The server expects the PORT environment variable to contain the port to run on (eg: 8000).

* **DATABASE_URL** The server expects the DATABASE_URL environment variable to contain the dataSourceName.  See [init.sql](sql/init.sql).

### Compile and run server

* **Install** The server can be compiled with `go install`.  The installed binary can be run with `$GOPATH/bin/nate-mlb`.

* **1-Command** To set environment variables, compile, and run the server with one command, run the command below and open a browser to http://<SERVER_HOST>:<SERVER_PORT> (eg: http://localhost:8000)
```
PORT=<SERVER_PORT> DATABASE_URL=<DATA_SOURCE_NAME> go run main.go
```

* **Heroku** To run locally with the [Heroku CLI](https://github.com/heroku/cli), create an `.env` file in the project root.  It should contain the environment variables on separate lines.  Run with `heroku local`.  Example .env file: 
```
PORT=<SERVER_PORT>
DATABASE_URL=<DATA_SOURCE_NAME>
```
