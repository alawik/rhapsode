package main

import (
    "fmt"
    "os"
    "os/exec"
    "strings"
)

func isDockerInstalled() bool {
    if _, err := os.Stat("/usr/local/bin/docker"); os.IsNotExist(err) {
        return false
    } else {
        return true
    }
}

func DockerBuild(fxPath string, imageTag string) {
    if isDockerInstalled() {
        cmd := "docker"
        args := []string{"build", "-t", imageTag, fxPath}
        if err := exec.Command(cmd, args...).Run(); err != nil {
            fmt.Fprintln(os.Stderr, err, "\nMake sure Docker is running.")
            os.Exit(1)
        }

        fmt.Printf("Docker image of function built.\n")
    } else {
        fmt.Print("Docker not installed.")
        os.Exit(1)
    }
}

func DockerRun(imageTag string, port string) {
    cmd := "docker"
    p := []string{port, ":8080"}
    ports := strings.Join(p, "")
    args := []string{"run", "-p", ports, "-d", imageTag}
    if err := exec.Command(cmd, args...).Run(); err != nil {
        fmt.Fprintln(os.Stderr, err, "\nMake sure Docker is running.")
        os.Exit(1)
    }

    fmt.Printf("Function running on port %s.\n", port)
}
