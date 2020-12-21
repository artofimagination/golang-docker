package docker

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/jhoonb/archivex"
	"github.com/pkg/errors"
)

func parseResponse(reader io.Reader) (map[string]interface{}, error) {
	d := json.NewDecoder(reader)
	result := make(map[string]interface{})
	for {
		if err := d.Decode(&result); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
	}
	return result, nil
}

func StartContainer(ID string) error {
	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}

	if err := cli.ContainerStart(context.Background(), ID, types.ContainerStartOptions{}); err != nil {
		err = errors.Wrap(errors.WithStack(err), "Failed to start container")
		return err
	}
	return nil
}

func tarballFolder(contextName string, sourceDirectory string) error {
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	defer tw.Close()

	tar := new(archivex.TarFile)
	if err := tar.Create(contextName); err != nil {
		return err
	}
	if err := tar.AddAll(sourceDirectory, false); err != nil {
		return err
	}
	if err := tar.Close(); err != nil {
		return err
	}

	return nil
}

func CreateImage(filePath string, imageName string) error {
	log.Println(filePath)
	log.Println(imageName)
	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}

	contextName := "context.tar"
	if err := tarballFolder(contextName, filePath); err != nil {
		return err
	}
	dockerBuildContext, err := os.Open(contextName)
	if err != nil {
		return err
	}

	imageBuildResponse, err := cli.ImageBuild(
		context.Background(),
		dockerBuildContext,
		types.ImageBuildOptions{
			Context:    dockerBuildContext,
			Dockerfile: "Dockerfile",
			Tags:       []string{imageName},
			Remove:     true})
	if err != nil {
		if errContext := dockerBuildContext.Close(); errContext != nil {
			return errors.Wrap(errors.WithStack(err), errContext.Error())
		}
		return err
	}
	defer imageBuildResponse.Body.Close()

	result, err := parseResponse(imageBuildResponse.Body)
	if err != nil {
		if errContext := dockerBuildContext.Close(); errContext != nil {
			return errors.Wrap(errors.WithStack(err), errContext.Error())
		}
		return err
	}
	value, ok := result["errorDetail"]
	if ok {
		dockerError := errors.New(value.(map[string]interface{})["message"].(string))
		if errContext := dockerBuildContext.Close(); errContext != nil {
			return errors.Wrap(errors.WithStack(dockerError), errContext.Error())
		}
		log.Println(dockerError)
		return dockerError
	}

	return nil
}

// CreateNewContainer creates and starts a docker container using an existing image
// defined by imageName
func CreateNewContainer(imageName string, address string, port string) (string, error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		err = fmt.Errorf("Unable to create docker client: %s", err.Error())
		return "", err
	}

	hostBinding := nat.PortBinding{
		HostIP:   address,
		HostPort: port,
	}
	containerPort, err := nat.NewPort("tcp", port)
	if err != nil {
		err = fmt.Errorf("Failed to get port: %s", err.Error())
		return "", err
	}

	portBinding := nat.PortMap{containerPort: []nat.PortBinding{hostBinding}}
	cont, err := cli.ContainerCreate(
		context.Background(),
		&container.Config{
			Image: imageName,
		},
		&container.HostConfig{
			PortBindings: portBinding,
		}, nil, "")
	if err != nil {
		err = fmt.Errorf("Failed to create docker container: %s", err.Error())
		return "", err
	}

	return cont.ID, nil
}
