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
    } else if nArgs >= 3 {
        fxPath := os.Args[2]
        fxPathExists, _ := DoesPathExist(fxPath)
        port := os.Args[3]

        if fxPathExists && port != "" {
            //fxLang := LangOfFunc(fxPath)

            //imageTag:=DockerBuild(fxPath)
            //DockerRun(imageTag, port)

            fmt.Print("Deployment complete.\n")
            os.Exit(0)
        } else {
            fmt.Print("Function path does not exist or image tag not defined or port not defined.")
            os.Exit(1)
            // If this happens, a runtime error will occur so this information wont be printed.
        }
    } else {
        fmt.Print("Unknown error.")
        os.Exit(1)
    }
}
