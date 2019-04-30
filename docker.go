package main

import (
	"archive/tar"
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/boltdb/bolt"
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

const dbDockerFile = "docker.db"
const imagesBucket = "images"

type DockerDB struct {
	db *bolt.DB
}
type DockerImage struct {
	Tag   string
	Value string
}

func getDockerDB() DockerDB {
	dockerDB := DockerDB{}
	dockerDB.Open()
	return dockerDB
}
func (ddb *DockerDB) Open() {
	db, err := bolt.Open(dbDockerFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}
	ddb.db = db
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(imagesBucket))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}
func (ddb *DockerDB) GetImage(tag string) (DockerImage, error) {
	var value []byte
	err := ddb.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(imagesBucket))
		value = b.Get([]byte(tag))
		return nil
	})
	if err != nil {
		log.Panic(err)
		return DockerImage{}, err
	}
	return DockerImage{
		tag,
		string(value),
	}, nil
}
func (ddb *DockerDB) AddImage(tag string, value string) (bool, error) {
	_, err := ddb.GetImage(tag)
	if err == nil {
		return true, nil
	}
	err = ddb.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(imagesBucket))
		err := b.Put([]byte(tag), []byte(value))
		return err
	})
	if err != nil {
		log.Panic(err)
		return false, err
	}
	return true, nil
}
func (ddb *DockerDB) DeleteImage(tag string) (bool, error) {
	err := ddb.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(imagesBucket))
		err := b.Delete([]byte(tag))
		return err
	})
	if err != nil {
		log.Panic(err)
		return false, err
	}
	return true, nil
}
func (ddb *DockerDB) GetImageList() ([]DockerImage, error) {
	var list []DockerImage
	err := ddb.db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte(imagesBucket))
		err := b.ForEach(func(k, v []byte) error {
			list = append(list, DockerImage{
				string(k),
				string(k),
			})
			return nil
		})
		if err != nil {
			log.Panic(err)
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
		return []DockerImage{}, err
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
	db := getDockerDB()
	_, err := db.AddImage(imageTag, imageTag)
	if err != nil {
		log.Panic("ERROR ADD IMAGE TO DB", err)
	}
	return imageTag
}
func DockerRun(imageTag string, port string) {
	db := getDockerDB()
	image, err := db.GetImage(imageTag)
	if err != nil {
		log.Panic("ERROR GET IMAGE FROM DB", err)
	}
	ctx := context.Background()
	cli := getDockerClient()
	//portSet:=nat.PortSet{"8080": struct{}{}}
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: image.Tag,
		//ExposedPorts: portSet,
	}, &container.HostConfig{
		//PortBindings: map[nat.Port][]nat.PortBinding{nat.Port(port): {{HostIP: "127.0.0.1", HostPort: port}}},
	}, nil, image.Tag)
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
func DockerImageList() {
	db := getDockerDB()
	list, err := db.GetImageList()
	if err != nil {
		log.Panic("ERROR GET IMAGE LIST FROM DB", err)
	}
	println("IMAGE LIST:")
	for i := range list {
		println(fmt.Sprintf("%v. %v - %v", i, list[i].Tag, list[i].Value))
	}
}
func DockerDeleteImage(tag string) {
	db := getDockerDB()
	_, err := db.DeleteImage(tag)
	if err != nil {
		log.Panic("ERROR DELETE IMAGE FROM DB", err)
		return
	}
	println(fmt.Sprintf("IMAGE %v WAS DELETED", tag))
}
