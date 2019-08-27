# nate-mlb
A web server which compares MLB baseball scores and NFL football scores.
Built with the [Go](https://github.com/golang/go) programming language.  Runs on a [Postgresql](https://github.com/postgres/postgres) database.


# Dependencies
* [pq](https://github.com/lib/pq) (Postgresql Driver)
* [Bootstrap](https://github.com/twbs/bootstrap) (css, html widgets)
* [jQuery](https://github.com/jquery/jquery) (Javascript functions for Bootstrap modals)
* [Font-Awesome](https://github.com/FortAwesome/Font-Awesome) (icons on about page)
* [bcrypt](https://github.com/golang/crypto) (password encryption)
* [godep](Godeps/Readme) dependency tool for Go. -> After altering dependencies, [run godep to update Heroku](https://devcenter.heroku.com/articles/go-dependencies-via-godep).


# Installation
### Create the the Postgresql Database.
See [init.sql](sql/init.sql) for setup instructions.
### Set environment variables.
The server expects a PORT environment variable to run on (eg: 8000) DATABASE_URL with dataSourceName (see [init.sql](sql/init.sql)).
### Compile and run the server.
* The server can be compiled with `go install` and run with `$GOPATH/bin/nate-mlb`.
* To set environment variables, compile, and run the server with one command, run the command below and open a browser to http://<SERVER_HOST>:<SERVER_PORT> (eg: http://localhost:8000)
```
PORT=<SERVER_PORT> DATABASE_URL=<DATA_SOURCE_NAME> go run main.go
```
* To run locally with the [Heroku CLI](https://github.com/heroku/cli), create a `.env` file next in the project root with the PORT AND DATABASE_URL properties and run `heroku local`.  Example .env file contents: 
```
PORT=<SERVER_PORT>
DATABASE_URL=<DATA_SOURCE_NAME>
```