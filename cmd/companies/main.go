package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/vkalekis/companies/internal"
	"github.com/vkalekis/companies/internal/config"
)

func main() {
	log := log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)

	confFile := flag.String("conf", "", "location of the yaml conf file")
	flag.Parse()

	if *confFile == "" || confFile == nil {
		log.Printf("Empty conf argument")
		return
	}

	config, err := config.ReadConfig(*confFile)
	if err != nil {
		log.Printf("Error during config retrieval: %v", err)
		return
	}
	config.Validate()

	app, err := internal.NewApp(config, log)
	if err != nil {
		log.Fatalf("Error initializing app: %v", err)
	}
	if err := app.Start(); err != nil {
		log.Fatalf("Error starting app: %v", err)
	}

	log.Printf("Companies app started")

	osSigCh := make(chan os.Signal, 1)
	signal.Notify(osSigCh, syscall.SIGINT, syscall.SIGTERM)
	<-osSigCh

	app.Stop()
	log.Printf("Companies app stopped")
}
