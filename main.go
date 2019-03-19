package main

import (
    "fmt"
    "os"
)

const version string = "0.0.0"

const usage = `Usage:
    serve      starts the server
    version    show current version
    help       show this information
`

func main() {
    if len(os.Args) < 2 {
        fmt.Print(usage)
        os.Exit(1)
    }

    switch os.Args[1] {
    case "version":
        fmt.Print(version)
        os.Exit(0)
    case "help":
        fmt.Print(usage)
        os.Exit(0)
    case "serve":
        Serve()
    default:
        fmt.Print(usage)
        os.Exit(1)
    }
}
