package main

import (
	"sync"

	"github.com/ccgg1997/Go-ZincSearch/api"
	_ "github.com/ccgg1997/Go-ZincSearch/docs"
	customHTTP "github.com/ccgg1997/Go-ZincSearch/modules/email/http"
)

func main() {
	var waitGr sync.WaitGroup
	var emailHandler *customHTTP.EmailHandler

	//instance the ingestion script and the pprof
	waitGr.Add(1)
	go handleIngestion(&waitGr)
	//intance the email module
	emailHandler = instanceEmail()
	waitGr.Wait()

	//instance the server
	mux := api.Routes(emailHandler)
	server := api.NewServer(mux)
	server.Run()
}
