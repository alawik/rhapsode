package main

import (
	"archive/tar"
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
)
import "github.com/mholt/archiver"

func isDockerInstalled() bool {
	if _, err := os.Stat("/usr/local/bin/docker"); os.IsNotExist(err) {
		return false
	} else {
		return true
	}
}
func getDockerClient() *client.Client {
	cli, err := client.NewClientWithOpts(client.WithVersion("1.39"))
	if err != nil {
		log.Fatal(err, " :unable to init client")
	}
	return cli
}
func checkIsFileExists(fileName string) bool {
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		return false
	}
	return true
}
func removeIfExists(fileName string) {
	log.Println("check if exists", fileName)
	if checkIsFileExists(fileName) {
		log.Println("remove", fileName)
		_ = os.Remove(fileName)
	}
}
func createBuildTar(dirPath string) string {
	log.Println("create build.tar")
	fileList, _ := ioutil.ReadDir(dirPath)
	fileListNames := make([]string, len(fileList))
	for i, file := range fileList {
		fileListNames[i] = strings.Join([]string{dirPath, file.Name()}, string(os.PathSeparator))
		log.Println(strings.Join([]string{dirPath, file.Name()}, string(os.PathSeparator)))
	}
	tarPath := strings.Join([]string{dirPath, "build.tar"}, string(os.PathSeparator))
	removeIfExists(tarPath)
	err := archiver.Archive(fileListNames, tarPath)
	if err != nil {
		log.Printf("Error in taring the docker root folder - %s", err.Error())
	}
	return tarPath
}
func writeToLog(reader io.ReadCloser) error {
	defer reader.Close()
	rd := bufio.NewReader(reader)
	for {
		n, _, err := rd.ReadLine()
		if err != nil && err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		log.Println(string(n))
	}
	return nil
}
func DockerBuild(fxPath string) string {
	imageTag := GenerateFunctionId()
	if isDockerInstalled() {
		ctx := context.Background()
		cli := getDockerClient()

		buf := new(bytes.Buffer)
		tw := tar.NewWriter(buf)
		defer tw.Close()

		dockerFile := "Dockerfile"
		dockerFileReader, err := os.Open(strings.Join([]string{fxPath, dockerFile}, string(os.PathSeparator)))
		if err != nil {
			log.Fatal(err, " :unable to open Dockerfile")
		}
		readDockerFile, err := ioutil.ReadAll(dockerFileReader)
		if err != nil {
			log.Fatal(err, " :unable to read dockerfile")
		}

		tarHeader := &tar.Header{
			Name: dockerFile,
			Size: int64(len(readDockerFile)),
		}
		err = tw.WriteHeader(tarHeader)
		if err != nil {
			log.Fatal(err, " :unable to write tar header")
		}
		_, err = tw.Write(readDockerFile)
		if err != nil {
			log.Fatal(err, " :unable to write tar body")
		}

		tarPath := createBuildTar(fxPath)

		dockerFileTarReader, _ := os.Open(tarPath)
		//buildArgs := make(map[string]*string)
		// add any build args if you want to
		//buildArgs["ENV"] = os.Getenv("GO_ENV")

		imageBuildResponse, err := cli.ImageBuild(
			ctx,
			dockerFileTarReader,
			types.ImageBuildOptions{
				NoCache:    true,
				Remove:     true,
				Tags:       []string{imageTag},
				Dockerfile: dockerFile,
				//BuildArgs:  buildArgs,
			})
		if err != nil {
			log.Fatal(err, " :unable to build docker image")
		}
		writeToLog(imageBuildResponse.Body)
		defer func() {
			imageBuildResponse.Body.Close()
			removeIfExists(tarPath)
		}()
		//_, err = io.Copy(os.Stdout, imageBuildResponse.Body)
		//if err != nil {
		//	log.Fatal(err, " :unable to read image build response")
		//}

	} else {
		fmt.Print("Docker not installed.")
		os.Exit(1)
	}
	log.Println("imageTag", imageTag)
	return imageTag
}

func DockerRun(imageTag string, port string) {

	ctx := context.Background()
	cli := getDockerClient()
	//portSet:=nat.PortSet{"8080": struct{}{}}
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: imageTag,
		//ExposedPorts: portSet,
	}, &container.HostConfig{
		//PortBindings: map[nat.Port][]nat.PortBinding{nat.Port(port): {{HostIP: "127.0.0.1", HostPort: port}}},
	}, nil, imageTag)
	if err != nil {
		panic(err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	//ctx := context.Background()
	//cli, err := client.NewEnvClient()
	//if err != nil {
	//	panic(err)
	//}
	//
	////reader, err := cli.ImagePull(ctx, "docker.io/library/alpine", types.ImagePullOptions{})
	////if err != nil {
	////    panic(err)
	////}
	////io.Copy(os.Stdout, reader)
	////
	////resp, err := cli.ContainerCreate(ctx, &container.Config{
	////    Image: "alpine",
	////    Cmd:   []string{"echo", "hello world"},
	////    Tty:   true,
	////}, nil, nil, "")
	////if err != nil {
	////    panic(err)
	////}
	//
	//if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
	//	panic(err)
	//}
	//
	//statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	//select {
	//case err := <-errCh:
	//	if err != nil {
	//		panic(err)
	//	}
	//case <-statusCh:
	//}
	//
	//out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
	//if err != nil {
	//	panic(err)
	//}
	//
	//io.Copy(os.Stdout, out)
	//
	//cmd := "docker"
	//p := []string{port, ":8080"}
	//ports := strings.Join(p, "")
	//args := []string{"run", "-p", ports, "-d", imageTag}
	//if err := exec.Command(cmd, args...).Run(); err != nil {
	//	fmt.Fprintln(os.Stderr, err, "\nMake sure Docker is running.")
	//	os.Exit(1)
	//}
	//
	log.Printf("Function running on port %s.\n", port)
}
