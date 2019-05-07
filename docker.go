package main

import (
	"archive/tar"
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/phayes/freeport"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
)
import "github.com/mholt/archiver"

const dbDockerFile = "functions.db"
const functionsBucket = "functions"

type DockerDB struct {
	db     *bolt.DB
	isOpen bool
}
type DockerFunction struct {
	Id          string `json:"id"`
	ContainerId string `json:"containerId"`
	IsRunning   bool   `json:"isRunning"`
	Port        int    `json:"port"`
	IP          string `json:"ip"`
}

var DockerDBInstance DockerDB

func getDockerDB() DockerDB {
	DockerDBInstance.Open()
	return DockerDBInstance
}
func (ddb *DockerDB) Open() {
	if ddb.isOpen {
		return
	}
	db, err := bolt.Open(dbDockerFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}
	ddb.isOpen = true
	ddb.db = db
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(functionsBucket))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}
func (ddb *DockerDB) GetFunction(functionID string) (DockerFunction, error) {
	var value []byte
	var dockerFunction DockerFunction
	err := ddb.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(functionsBucket))
		value = b.Get([]byte(functionID))

		jsonErr := json.Unmarshal(value, &dockerFunction)
		if jsonErr != nil {
			log.Panic("Parse function JSON error", jsonErr)
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
		return dockerFunction, err
	}
	return dockerFunction, nil
}
func (ddb *DockerDB) AddFunction(functionID string, containerID string) (bool, error) {
	err := ddb.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(functionsBucket))
		jsonBytes, jsonErr := json.Marshal(DockerFunction{
			Id:          functionID,
			ContainerId: containerID,
			IsRunning:   false,
			Port:        0,
			IP:          "",
		})
		if jsonErr != nil {
			log.Panic("Stringify function JSON error", jsonErr)
		}
		err := b.Put([]byte(functionID), jsonBytes)
		return err
	})
	if err != nil {
		log.Panic(err)
		return false, err
	}
	return true, nil
}
func (ddb *DockerDB) UpdateFunction(dockerFunction DockerFunction) (bool, error) {
	err := ddb.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(functionsBucket))
		jsonBytes, jsonErr := json.Marshal(dockerFunction)
		if jsonErr != nil {
			log.Panic("Stringify function JSON error", jsonErr)
		}
		err := b.Put([]byte(dockerFunction.Id), jsonBytes)
		return err
	})
	if err != nil {
		log.Panic(err)
		return false, err
	}
	return true, nil
}
func (ddb *DockerDB) DeleteFunction(functionID string) (bool, error) {
	err := ddb.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(functionsBucket))
		err := b.Delete([]byte(functionID))
		return err
	})
	if err != nil {
		log.Panic(err)
		return false, err
	}
	return true, nil
}
func (ddb *DockerDB) GetFunctionList() ([]DockerFunction, error) {
	var list []DockerFunction
	err := ddb.db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte(functionsBucket))
		err := b.ForEach(func(k, v []byte) error {
			dockerFunction := DockerFunction{}
			jsonErr := json.Unmarshal(v, &dockerFunction)
			if jsonErr != nil {
				log.Panic("Parse function JSON error", jsonErr)
			}
			list = append(list, dockerFunction)
			return nil
		})
		if err != nil {
			log.Panic(err)
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
		return []DockerFunction{}, err
	}
	return list, nil
}

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
	fxPathExists, _ := DoesPathExist(fxPath)
	if !fxPathExists {
		fmt.Print("Function path does not exist.")
		os.Exit(1)
	}
	functionID := GenerateFunctionId()
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
				Tags:       []string{functionID},
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
			log.Println(fmt.Sprintf("Create container from image %v", functionID))
			resp, err := cli.ContainerCreate(ctx, &container.Config{Image: functionID}, &container.HostConfig{}, nil, functionID)
			if err != nil {
				panic(err)
			}
			log.Println(fmt.Sprintf("Container ID = %v", resp.ID))
			db := getDockerDB()
			_, err = db.AddFunction(functionID, resp.ID)
			log.Println(fmt.Sprintf("Function %v added to DB with contaienr ID = %v", functionID, resp.ID))
			if err != nil {
				log.Panic("ERROR ADD FUNCTION TO DB", err)
			}
		}()
		//_, err = io.Copy(os.Stdout, imageBuildResponse.Body)
		//if err != nil {
		//	log.Fatal(err, " :unable to read image build response")
		//}

	} else {
		fmt.Print("Docker not installed.")
		os.Exit(1)
	}
	log.Println("functionID", functionID)

	return functionID
}
func DockerRun(functionID string) {
	port, err := freeport.GetFreePort()
	if err != nil {
		log.Panic("Can't find free port", err)
	}

	db := getDockerDB()
	function, err := db.GetFunction(functionID)
	if err != nil {
		log.Panic("ERROR GET FUNCTION FROM DB", err)
		os.Exit(1)
	}

	ctx := context.Background()
	cli := getDockerClient()
	//portSet:=nat.PortSet{"8080": struct{}{}}

	_ = cli.ContainerStop(ctx, function.ContainerId, nil)
	log.Println(fmt.Sprintf("Runing function %v (docker container %v)", function.Id, function.ContainerId))
	if err := cli.ContainerStart(ctx, function.ContainerId, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}
	function.Port = port
	function.IsRunning = true
	function.IP = string("127.0.0.1")
	ok, err := db.UpdateFunction(function)
	if !ok || err != nil {
		log.Panic("Err DB update function", err)
	}

	log.Printf("Function running on port %v.\n", port)
}

func DockerStop(containerId string) {
	ctx := context.Background()
	cli := getDockerClient()
	err := cli.ContainerStop(ctx, containerId, nil)
	if err != nil {
		panic(err)
	}

	log.Printf("Function with ID %v has been stopped", containerId)
}
func DockerRemoveContainer(containerId string) {
	ctx := context.Background()
	cli := getDockerClient()
	err := cli.ContainerRemove(ctx, containerId, types.ContainerRemoveOptions{})
	if err != nil {
		panic(err)
	}
	log.Printf("Container with ID %v has been removed", containerId)
}
func DockerRemoveImage(functionID string) {
	ctx := context.Background()
	cli := getDockerClient()
	image, _, err := cli.ImageInspectWithRaw(ctx, functionID)
	if err != nil {
		panic(err)
	}
	if image.ID != "" {
		_, err = cli.ImageRemove(ctx, image.ID, types.ImageRemoveOptions{})
	}
	if err != nil {
		panic(err)
	}
	log.Printf("Image with ID %v has been remmoved", functionID)
}
func DockerFunctionList() {
	db := getDockerDB()
	list, err := db.GetFunctionList()
	if err != nil {
		log.Panic("ERROR GET FUNCTION LIST FROM DB", err)
	}
	if len(list) == 0 {
		println("Function list is empty")
		os.Exit(0)
	}
	println(fmt.Sprintf("Functions (%v):", len(list)))
	println(fmt.Sprintf("   ID                     Container ID"))
	for i := range list {
		status := ""
		if list[i].IsRunning {
			status = "running"
		} else {
			status = "not running"
		}
		ipString := ""
		if list[i].IP != "" && list[i].Port > 0 {
			ipString = fmt.Sprintf(" (%v:%v)", list[i].IP, list[i].Port)
		}
		println(fmt.Sprintf("%v. %v - %v%v %v", i, list[i].Id, list[i].ContainerId, ipString, status))
	}
}
func DockerDeleteFunction(functionID string) {
	db := getDockerDB()
	function, err := db.GetFunction(functionID)
	if err != nil {
		log.Println("ERROR GET FUNCTION FROM DB", err)
		os.Exit(1)
	}
	if function.Id == "" {
		log.Println(fmt.Sprintf("Function %v not found", functionID))
		os.Exit(1)
	}
	if function.ContainerId == "" {
		log.Println(fmt.Sprintf("Container not found"))
		os.Exit(1)
	}

	DockerStop(function.ContainerId)
	DockerRemoveContainer(function.ContainerId)
	DockerRemoveImage(function.Id)
	_, err = db.DeleteFunction(functionID)
	if err != nil {
		log.Panic("ERROR DELETE FUNCTION FROM DB", err)
		return
	}
	log.Printf("Function %v has been deleted", functionID)
}
func DockerStopFunction(functionID string) {
	db := getDockerDB()
	function, err := db.GetFunction(functionID)
	if err != nil {
		log.Println("Error get function from db", err)
		os.Exit(1)
	}
	if function.Id == "" {
		log.Println(fmt.Sprintf("Function %v not found", functionID))
		os.Exit(1)
	}
	if function.ContainerId == "" {
		log.Println(fmt.Sprintf("Container not found"))
		os.Exit(1)
	}

	DockerStop(function.ContainerId)
	function.IsRunning = false
	function.Port = 0
	function.IP = ""
	ok, err := db.UpdateFunction(function)
	if !ok || err != nil {
		log.Panic("Err DB update function", err)
	}
}
