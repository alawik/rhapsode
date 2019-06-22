# Rhapsode

[![License](https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)](https://github.com/alawik/rhapsode/blob/master/LICENSE)
[![Version](https://img.shields.io/badge/version-0.0.0-lightgrey.svg?style=flat-square)](https://github.com/alawik/rhapsode)

Rhapsode is a distributed function-as-a-service platform that utilizes blockchain and synchronization technology to enable developers the freedom of executing their deployed code on any node within the network.

Work In Progress (WIP)

## Features

```
Usage:
  Blockchain:
    createblockchain -address ADDRESS - Create a blockchain and send genesis block reward to ADDRESS
    createwallet - Generates a new key-pair and saves it into the wallet file
    getbalance -address ADDRESS - Get balance of ADDRESS
    listaddresses - Lists all addresses from the wallet file
    printchain - Print all the blocks of the blockchain
    reindexutxo - Rebuilds the UTXO set
    send -from FROM -to TO -amount AMOUNT - Send AMOUNT of coins from FROM address to TO
  Server:
    serve - Starts the server

  Functions:
    fbuild - Build function to image and add to function list
    frun - Run function by id
    fstop - Stop function by id
    flist - Get list of functions
    fdelete - Remove function by id

  General:
    version - Shows the current software version
    help - Show this usage information
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

[github.com/gorilla/mux](https://github.com/gorilla/mux) <br>
[github.com/boltdb/bolt](https://github.com/boltdb/bolt) <br>
[golang.org/x/crypto/ripemd160](https://golang.org/x/crypto/ripemd160)

## License

MIT
