# Websocket Demo

This is a quick demo to share how a multi-room websocket application could be designed with Go

## Pre-requirements

CGO is needed in order for this application to run, therefore please be sure that has been setup.

Starting the server is as simple as `go run .`. There are flags for **port** and **sqlite file location** which are `-p` and `-f` respectively. These are optional.

The migration file needs to be ran manually. I would advise to use the [sqlite3 cli](https://sqlite.org/cli.html#:~:text=Start%20the%20sqlite3%20program%20by,name%20will%20be%20created%20automatically.) to achieve this.

## Todo

Plans to create a docker image in the future.
