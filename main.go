package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/artofimagination/golang-docker/docker"
	"github.com/pkg/errors"

	"github.com/gorilla/mux"
)

func helloServer(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, I'm test server!")
}

const (
	POST = "POST"
	GET  = "GET"
)

func checkRequestType(requestTypeString string, w http.ResponseWriter, r *http.Request) error {
	if r.Method != requestTypeString {
		w.WriteHeader(http.StatusBadRequest)
		errorString := fmt.Sprintf("Invalid request type %s", r.Method)
		fmt.Fprint(w, errorString)
		return errors.New(errorString)
	}
	return nil
}

func decodePostData(w http.ResponseWriter, r *http.Request) (map[string]interface{}, error) {
	if err := checkRequestType(POST, w, r); err != nil {
		return nil, err
	}

	data := make(map[string]interface{})
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		err = errors.Wrap(errors.WithStack(err), "Failed to decode request json")
		return nil, err
	}

	return data, nil
}

func getImage(w http.ResponseWriter, r *http.Request) {
	log.Println("Getting Image")
	if err := checkRequestType(GET, w, r); err != nil {
		return
	}

	names, ok := r.URL.Query()["image-name"]
	if !ok || len(names[0]) < 1 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, errors.New("Url Param 'image-name' is missing"))
		return
	}

	images, err := docker.ListImages()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err.Error())
		return
	}

	_, err = docker.GetImageIDByTag(images, names[0])
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, names[0])
}

func createImage(w http.ResponseWriter, r *http.Request) {
	log.Println("Creating Image")
	data, err := decodePostData(w, r)
	if err != nil {
		return
	}

	name, ok := data["image-name"].(string)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Missing 'image-name'")
		return
	}

	source, ok := data["source-dir"].(string)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Missing 'source-dir'")
		return
	}

	if err := docker.CreateImage(source, name); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err.Error())
		return
	}

	images, err := docker.ListImages()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err.Error())
		return
	}

	_, err = docker.GetImageIDByTag(images, name)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, name)
}

func deleteImage(w http.ResponseWriter, r *http.Request) {
	log.Println("Deleting Image")
	data, err := decodePostData(w, r)
	if err != nil {
		return
	}

	name, ok := data["image-name"].(string)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Missing 'image-name'")
		return
	}

	images, err := docker.ListImages()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err.Error())
		return
	}

	ID, err := docker.GetImageIDByTag(images, name)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err.Error())
		return
	}

	if err := docker.DeleteImage(ID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err.Error())
		return
	}

	images, err = docker.ListImages()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err.Error())
		return
	}

	_, err = docker.GetImageIDByTag(images, name)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Delete completed")
		return
	}

	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprint(w, "Failed to delete")
}

func createContainer(w http.ResponseWriter, r *http.Request) {
	log.Println("Creating container")
	data, err := decodePostData(w, r)
	if err != nil {
		return
	}

	name, ok := data["image-name"].(string)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Missing 'image-name'")
		return
	}

	port, ok := data["port"].(string)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Missing 'port'")
		return
	}

	address, ok := data["address"].(string)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Missing 'address'")
		return
	}

	ID, err := docker.CreateNewContainer(name, address, port)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Container created: %s", ID)
}

func startContainer(w http.ResponseWriter, r *http.Request) {
	log.Println("Starting container")
	if err := checkRequestType(GET, w, r); err != nil {
		return
	}

	ids, ok := r.URL.Query()["id"]
	if !ok || len(ids[0]) < 1 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, errors.New("Url Param 'id' is missing"))
		return
	}

	if err := docker.StartContainer(ids[0]); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Container started")
}

func stopContainer(w http.ResponseWriter, r *http.Request) {
	log.Println("Stopping container")
	if err := checkRequestType(GET, w, r); err != nil {
		return
	}

	ids, ok := r.URL.Query()["id"]
	if !ok || len(ids[0]) < 1 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, errors.New("Url Param 'id' is missing"))
		return
	}

	if err := docker.StopContainer(ids[0]); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Container stopped")
}

func getImageIDByTag(w http.ResponseWriter, r *http.Request) {
	if err := checkRequestType(GET, w, r); err != nil {
		return
	}

	names, ok := r.URL.Query()["image-name"]
	if !ok || len(names[0]) < 1 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, errors.New("Url Param 'image-name' is missing"))
		return
	}

	images, err := docker.ListImages()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err.Error())
		return
	}

	ID, err := docker.GetImageIDByTag(images, names[0])
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, ID)
}

func stopContainerByImageID(w http.ResponseWriter, r *http.Request) {
	log.Println("Stopping container by image ID")
	if err := checkRequestType(GET, w, r); err != nil {
		return
	}

	ids, ok := r.URL.Query()["id"]
	if !ok || len(ids[0]) < 1 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, errors.New("Url Param 'id' is missing"))
		return
	}

	if err := docker.StopContainerByImageID(ids[0]); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Container stopped")
}

func deleteContainer(w http.ResponseWriter, r *http.Request) {
	log.Println("Deleting container")
	data, err := decodePostData(w, r)
	if err != nil {
		return
	}

	ID, ok := data["id"].(string)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Missing 'id'")
		return
	}

	if err := docker.DeleteContainer(ID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err.Error())
		return
	}

	containers, err := docker.ListContainers()
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}

	for _, container := range containers {
		if container.ID == ID {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "Container not deleted")
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Container deleted")
}

func containerExists(w http.ResponseWriter, r *http.Request) {
	log.Println("Container exists")
	data, err := decodePostData(w, r)
	if err != nil {
		return
	}

	ID, ok := data["id"].(string)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Missing 'id'")
		return
	}

	err = docker.ContainerExists(ID)
	if err == nil {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Container exists")
		return
	}

	if err == docker.ErrContainerNotFound {
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprint(w, err.Error())
		return
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err.Error())
		return
	}
}

func getContainer(w http.ResponseWriter, r *http.Request) {
	log.Println("Getting container")
	if err := checkRequestType(GET, w, r); err != nil {
		return
	}

	ids, ok := r.URL.Query()["id"]
	if !ok || len(ids[0]) < 1 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, errors.New("Url Param 'id' is missing"))
		return
	}

	containers, err := docker.ListContainers()
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}

	for _, container := range containers {
		if container.ID == ids[0] {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "Container found")
			return
		}
	}

	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprint(w, "Container not found")
}

func main() {

	r := mux.NewRouter()
	r.HandleFunc("/", helloServer)
	r.HandleFunc("/create-image", createImage)
	r.HandleFunc("/get-image", getImage)
	r.HandleFunc("/delete-image", deleteImage)
	r.HandleFunc("/get-image-id-by-tag", getImageIDByTag)
	r.HandleFunc("/create-container", createContainer)
	r.HandleFunc("/get-container", getContainer)
	r.HandleFunc("/start-container", startContainer)
	r.HandleFunc("/stop-container", stopContainer)
	r.HandleFunc("/stop-container-by-image-id", stopContainerByImageID)
	r.HandleFunc("/delete-container", deleteContainer)
	r.HandleFunc("/container-exists", containerExists)
	// Create Server and Route Handlers
	srv := &http.Server{
		Handler:      r,
		Addr:         ":8080",
		ReadTimeout:  20 * time.Second,
		WriteTimeout: 20 * time.Second,
	}

	// Start Server
	go func() {
		log.Println("Starting Server")
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	// Graceful Shutdown
	waitForShutdown(srv)
}

func waitForShutdown(srv *http.Server) {
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Block until we receive our signal.
	<-interruptChan

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}

	log.Println("Shutting down")
	os.Exit(0)
}
