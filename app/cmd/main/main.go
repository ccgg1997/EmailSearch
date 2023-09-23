package main

import (
	"fmt"
	_ "net/http/pprof"
	"os"
	"runtime/pprof"
	"sync"

	"github.com/ccgg1997/Go-ZincSearch/api"
	_ "github.com/ccgg1997/Go-ZincSearch/docs"
	"github.com/ccgg1997/Go-ZincSearch/email/gateway"
	customHTTP "github.com/ccgg1997/Go-ZincSearch/email/http"
	"github.com/ccgg1997/Go-ZincSearch/email/usecase"
	script "github.com/ccgg1997/Go-ZincSearch/script2"
)

func main() {
	// profiling
	cpuFile, err := os.Create("cpu.pprof")
	if err != nil {
		panic(err)
	}
	var waitGr sync.WaitGroup

	pprof.StartCPUProfile(cpuFile)

	waitGr.Add(1)
	go func() {
		ingesta := script.IngestaDeDatos()
		if ingesta {
			fmt.Println("<100> Se realizó la ingesta de datos")
		}

		fmt.Println("<101> Ya existen los datos. No se realizó la ingesta de datos")
		waitGr.Done()
	}()

	// Crear una instancia de EmailGateway
	emailGateway := gateway.NewEmailGateway("email")
	// Crear una instancia de EmailUsecase
	emailUsecase := usecase.NewEmailUsecase(*emailGateway)
	// Crear una instancia de EmailHandler
	emailHandler := customHTTP.NewEmailHandler(*emailUsecase)

	waitGr.Wait()
	fmt.Println("Finalizando...")
	// Forzar a vaciar el búfer de salida
	os.Stdout.Sync()
	pprof.StopCPUProfile()

	// Iniciar servidor web
	mux := api.Routes(emailHandler)
	server := api.NewServer(mux)
	server.Run()

}
