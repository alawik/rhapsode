package main

import (
    "fmt"
    "os"
)

func Deploy() {
    nArgs := len(os.Args)
    if nArgs == 2 {
        fmt.Print("Use the help command for usage.")
        os.Exit(1)
    } else if nArgs == 3 {
        fxPath := os.Args[2]
        fxPathExists, _ := DoesPathExist(fxPath)

        if fxPathExists {
            //fxLang := LangOfFunc(fxPath)

            var imageTag = "rtest/node_func:v2"
            DockerBuild(fxPath, imageTag)
            DockerRun(imageTag)

            fmt.Printf("Deployment complete.\n")
            os.Exit(0)
        } else {
            fmt.Print("Function path does not exist.")
            os.Exit(1)
        }
    } else {
        fmt.Print("Unknown error.")
        os.Exit(1)
    }
}
