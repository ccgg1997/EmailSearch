package main

import (
	"fmt"
	"sync"

	"github.com/ccgg1997/Go-ZincSearch/cmd/main/profiler"
	"github.com/ccgg1997/Go-ZincSearch/modules/email/gateway"
	httpModule "github.com/ccgg1997/Go-ZincSearch/modules/email/http"
	"github.com/ccgg1997/Go-ZincSearch/modules/email/usecase"
	script "github.com/ccgg1997/Go-ZincSearch/script/script3"
)

func instanceEmail() *httpModule.EmailHandler {
	//instanciar el gateway, el usecase y el handler de email
	emailGateway := gateway.NewEmailGateway("email")
	emailUsecase := usecase.NewEmailUsecase(emailGateway)
	emailHandler := httpModule.NewEmailHandler(emailUsecase)
	return emailHandler
}

func handleIngestion(waitGr *sync.WaitGroup) error {
	//instance the ingestion script and the pprof

	if err := profiler.StartCPUProfile(); err != nil {
		fmt.Printf("Error starting CPU profile: %v", err)
		return err
	}

	if err := profiler.StartHeapProfile(); err != nil {
		fmt.Printf("Error starting heap profile: %v", err)
		return err
	}
	
	ingesta, err := script.IngestaDeDatos()

	if err == nil && !ingesta {
		fmt.Println("Data already exists. Data ingestion skipped")
	}
	if err != nil {
		fmt.Println(err)
		fmt.Println("Data ingestion failed")
	}
	if err := profiler.StopCPUProfile(); err != nil {
		return err
	}
	profiler.StopHeapProfile()

	waitGr.Done()
	fmt.Print("Data ingestion succeeded")
	return nil
}
