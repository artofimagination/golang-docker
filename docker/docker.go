package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/jhoonb/archivex"
	"github.com/pkg/errors"
)

var ErrNoImagesDeleted = errors.New("No images were deleted")
var ErrImageNotFound = errors.New("Image not found")
var ErrContainerNotFound = errors.New("Container not found")

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

func tarballFolder(contextName string, sourceDirectory string) error {
	tar := new(archivex.TarFile)
	if err := tar.Create(contextName); err != nil {
		if errClose := tar.Close(); errClose != nil {
			return errors.Wrap(errors.WithStack(err), errClose.Error())
		}
		return err
	}
	if err := tar.AddAll(sourceDirectory, false); err != nil {
		if errClose := tar.Close(); errClose != nil {
			return errors.Wrap(errors.WithStack(err), errClose.Error())
		}
		return err
	}
	if err := tar.Close(); err != nil {
		return err
	}

	return nil
}

func DeleteImage(imageID string) error {
	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}

	imagesDeleted, err := cli.ImageRemove(context.Background(), imageID, types.ImageRemoveOptions{Force: true, PruneChildren: true})
	if err != nil {
		return err
	}

	if len(imagesDeleted) == 0 {
		return ErrNoImagesDeleted
	}
	return nil
}

type Images []types.ImageSummary

func GetImageIDByTag(i Images, inputTag string) (string, error) {
	for _, value := range i {
		for _, tag := range value.RepoTags {
			if tag == inputTag {
				return value.ID, nil
			}
		}
	}
	return "", ErrImageNotFound
}

func CreateImage(filePath string, imageName string) error {
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
		return dockerError
	}

	return nil
}

func ListImages() ([]types.ImageSummary, error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}

	images, err := cli.ImageList(context.Background(), types.ImageListOptions{})
	if err != nil {
		return nil, err
	}

	return images, nil
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

func DeleteContainer(ID string) error {
	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}

	if err := cli.ContainerRemove(
		context.Background(),
		ID,
		types.ContainerRemoveOptions{
			Force: true,
		}); err != nil {
		return err
	}
	return nil
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

func StopContainer(ID string) error {
	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}

	if err := cli.ContainerStop(context.Background(), ID, nil); err != nil {
		return err
	}
	return nil
}

func PauseContainer(ID string) error {
	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}

	if err := cli.ContainerPause(context.Background(), ID); err != nil {
		return err
	}
	return nil
}

func UnpauseContainer(ID string) error {
	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}

	if err := cli.ContainerUnpause(context.Background(), ID); err != nil {
		return err
	}
	return nil
}

func IsContainerRunning(ID string) bool {

	return true
}

func ListContainers() ([]types.Container, error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{All: true})
	if err != nil {
		return nil, err
	}

	return containers, nil
}

func ContainerExists(ID string) error {
	containerList, err := ListContainers()
	if err != nil {
		return err
	}

	for _, container := range containerList {
		if container.ID == ID {
			return nil
		}
	}

	return ErrContainerNotFound
}

func StopContainerByImageID(imageID string) error {
	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{All: true})
	if err != nil {
		return err
	}

	for _, container := range containers {
		if container.ImageID == imageID {
			if err := StopContainer(container.ID); err != nil {
				return err
			}
		}
	}
	return nil
}
