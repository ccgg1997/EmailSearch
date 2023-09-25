package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/ccgg1997/Go-ZincSearch/api"
	"github.com/ccgg1997/Go-ZincSearch/cmd/main/profiler"
	_ "github.com/ccgg1997/Go-ZincSearch/docs"
	"github.com/ccgg1997/Go-ZincSearch/email/gateway"
	customHTTP "github.com/ccgg1997/Go-ZincSearch/email/http"
	"github.com/ccgg1997/Go-ZincSearch/email/usecase"
	script "github.com/ccgg1997/Go-ZincSearch/script2"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}

// RUN IS THE REPONSIBLE OF STARTING THE APPLICATION
func run() error {
	var waitGr sync.WaitGroup

	if err := profiler.StartCPUProfile(); err != nil {
		return err
	}

	waitGr.Add(1)
	go func() {
		handleIngestion()
		waitGr.Done()
	}()

	emailGateway := gateway.NewEmailGateway("email")
	emailUsecase := usecase.NewEmailUsecase(emailGateway)
	emailHandler := customHTTP.NewEmailHandler(emailUsecase)

	waitGr.Wait()

	if err := profiler.StopCPUProfile(); err != nil {
		return err
	}

	mux := api.Routes(emailHandler)
	server := api.NewServer(mux)
	server.Run()
	return nil
}

func handleIngestion() {
	if script.IngestaDeDatos() {
		fmt.Println("Data ingestion succeeded")
	} else {
		fmt.Println("Data already exists. Data ingestion skipped")
	}
}
