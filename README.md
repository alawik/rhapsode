# Rhapsode

[![License](https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)](https://github.com/alawik/rhapsode/blob/master/LICENSE)
[![Version](https://img.shields.io/badge/version-0.0.0-lightgrey.svg?style=flat-square)](https://github.com/alawik/rhapsode)

Rhapsode is a distributed function-as-a-service platform that utilizes blockchain and synchronization technology to enable developers the freedom of executing their deployed code on any node within the network.

Work In Progress (WIP)

## Features

```
Usage:
    serve      starts the server
    deploy     builds and runs specified function in a Docker container
    version    show current version
    help       show this information
```

## Installation

`git clone https://github.com/alawik/rhapsode.git`

`dep ensure`

`go build`

### Usage

`./rhapsode deploy <function path> <port>`

`./rhapsode deploy ./functions/node/. 8080`

Currently there is no handler to send the deployment request to the server, but the function will still be deployed. Use the Docker CLI to stop and remove the running container.

## Dependencies

Docker should be installed on your machine.

[github.com/gorilla/mux](https://github.com/gorilla/mux)

## License

MIT
