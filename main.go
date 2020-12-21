package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/artofimagination/golang-docker/docker"

	"github.com/gorilla/mux"
)

func helloServer(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, I'm test server!")
}

func createImage(w http.ResponseWriter, r *http.Request) {
	log.Println("Creating Image")

	names, ok := r.URL.Query()["image-name"]
	if !ok || len(names[0]) < 1 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, errors.New("Url Param 'image-name' is missing"))
		return
	}

	directories, ok := r.URL.Query()["source-dir"]
	if !ok || len(directories[0]) < 1 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, errors.New("Url Param 'source-dir' is missing"))
		return
	}

	err := docker.CreateImage(directories[0], names[0])
	if err != nil {
		fmt.Fprintln(w, err.Error())
		return
	}
	fmt.Fprintln(w, "Image created")
}

func createContainer(w http.ResponseWriter, r *http.Request) {
	log.Println("Creating container")

	names, ok := r.URL.Query()["image-name"]
	if !ok || len(names[0]) < 1 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, errors.New("Url Param 'image-name' is missing"))
		return
	}

	ports, ok := r.URL.Query()["port"]
	if !ok || len(ports[0]) < 1 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, errors.New("Url Param 'port' is missing"))
		return
	}

	addresses, ok := r.URL.Query()["address"]
	if !ok || len(addresses[0]) < 1 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, errors.New("Url Param 'address' is missing"))
		return
	}

	id, err := docker.CreateNewContainer(names[0], addresses[0], ports[0])
	if err != nil {
		fmt.Fprintln(w, err.Error())
		return
	}
	fmt.Fprintln(w, "Container created")
	fmt.Fprintln(w, id)
}

func startContainer(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Starting container")

	ids, ok := r.URL.Query()["id"]
	if !ok || len(ids[0]) < 1 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, errors.New("Url Param 'id' is missing"))
		return
	}

	if err := docker.StartContainer(ids[0]); err != nil {
		fmt.Fprintln(w, err.Error())
		return
	}
	fmt.Fprintln(w, "Container started")
}

func main() {

	r := mux.NewRouter()
	r.HandleFunc("/", helloServer)
	r.HandleFunc("/create-image", createImage)
	r.HandleFunc("/create-container", createContainer)
	r.HandleFunc("/start-container", startContainer)
	// Create Server and Route Handlers
	srv := &http.Server{
		Handler:      r,
		Addr:         ":8081",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
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
