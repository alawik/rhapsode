package main

import (
    "fmt"
    "os"
    "os/exec"
)

func DockerBuild(fxPath string, imageTag string) {
    cmd := "docker"
    args := []string{"build", "-t", imageTag, fxPath}
    if err := exec.Command(cmd, args...).Run(); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }

    fmt.Printf("Docker image of function built.\n")
}

func DockerRun(imageTag string) {
    cmd := "docker"
    args := []string{"run", "-p", "8080:8080", "-d", imageTag}
    if err := exec.Command(cmd, args...).Run(); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }

    fmt.Printf("Function running on port 8080.\n")
}
