# ![nate-mlb favicon](static/favicon.ico) nate-mlb
A web server which compares MLB baseball scores and NFL football scores.

Built with the [Go](https://github.com/golang/go) programming language.

Runs on a [PostgreSQL](https://github.com/postgres/postgres) database.

[![Build Status](https://travis-ci.org/jacobpatterson1549/nate-mlb.svg?branch=master)](https://travis-ci.org/jacobpatterson1549/nate-mlb)
[![Go Report Card](https://goreportcard.com/badge/github.com/jacobpatterson1549/nate-mlb)](https://goreportcard.com/report/github.com/jacobpatterson1549/nate-mlb)
[![GoDoc](https://godoc.org/github.com/jacobpatterson1549/nate-mlb?status.svg)](https://godoc.org/github.com/jacobpatterson1549/nate-mlb)

## Dependencies
New dependencies are automatically added to [go.mod](go.mod) when the project is built.
* [pq](https://github.com/lib/pq) (PostgreSQL Driver)
* [bcrypt](https://github.com/golang/crypto) (password encryption)
* [Bootstrap](https://github.com/twbs/bootstrap) (css, html widgets)
* [Font-Awesome](https://github.com/FortAwesome/Font-Awesome) (icons on about page)


## Installation

### Database
The server expects to use PostgreSQL database.  See [Database Setup](sql/README.md) for instructions on creating the database.

### Set environment variables
The following environment variables should be set or provided:
* **PORT** The server expects the PORT environment variable to contain the port to run on (eg: 8000). **REQUIRED**
* **DATABASE_URL** The server expects the DATABASE_URL environment variable to contain the dataSourceName.  See [Database Setup](sql/README.md). **REQUIRED**
* **ADMIN_PASSWORD** The administrator password to edit years/players/friends on the site.
* **APPLICATION_NAME** The name of the application server to display to users  Visible on the site and on exports.
* **PLAYER_TYPES** A csv whitelist of [PlayerType](https://godoc.org/github.com/jacobpatterson1549/nate-mlb/go/db#PlayerType) ids to use.  If present, limits player types.  For example, when `4,5` is used, only player types nflTeam and nflQB will be shown; nfl will also be the only sport shown.

### Compile and run server
Two ways to compile and run the server are listed below.
* **Install** The server can be compiled with `go install`.  The installed binary can be run with `$GOPATH/bin/nate-mlb`.
* **1-Command** To set environment variables, compile, and run the server with one command, run the command below and open a browser to http://<SERVER_HOST>:<SERVER_PORT> (eg: http://localhost:8000)
```bash
# Any of these commands compile and run the server
go install && nate-mlb -p <PORT> -ds <DATA_SOURCE_NAME> -ap <ADMIN_PASSWORD>
go run main.go -p <PORT> -ds <DATA_SOURCE_NAME> -ap <ADMIN_PASSWORD>
PORT=<SERVER_PORT> DATABASE_URL=<DATA_SOURCE_NAME> ADMIN_PASSWORD=<ADMIN_PASSWORD> go run main.go
```

### Heroku
To run locally with the [Heroku CLI](https://github.com/heroku/cli), create an `.env` file in the project root.  It should contain the environment variables on separate lines.  Run with `heroku local`.  Example .env file: 
```bash
PORT=<SERVER_PORT>
DATABASE_URL=<DATA_SOURCE_NAME>
ADMIN_PASSWORD=<ADMIN_PASSWORD>
```
