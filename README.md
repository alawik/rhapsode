# Rhapsode

[![License](https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)](https://github.com/alawik/rhapsode/blob/master/LICENSE)
[![Version](https://img.shields.io/badge/version-0.0.0-lightgrey.svg?style=flat-square)](https://github.com/alawik/rhapsode)

Rhapsode is a distributed function-as-a-service platform that utilizes blockchain and synchronization technology to enable developers the freedom of executing their deployed code on any node within the network.

Work In Progress (WIP)

## Features

```
Usage:
    createwallet        generates new key-pair and saves it into the wallet file
    listaddresses       lists all addresses from the wallet file
    createblockchain    Create a blockchain and initiate genesis block 
    getbalance          get balance of an address
    printchain          prints all the blocks of the blockchain in terminal
    reindexutxo         rebuilds the UTXO set
    send                send coins from one address to another
    serve               starts the server
    deploy              builds and runs specified function in a Docker container
    version             show current version
    help                show this information
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
