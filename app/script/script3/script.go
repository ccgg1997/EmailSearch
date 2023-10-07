package script

import (
	"fmt"
	"sync"
	"time"

	zincClient "github.com/ccgg1997/Go-ZincSearch/internal/zincsearch"
	"github.com/ccgg1997/Go-ZincSearch/modules/email/models"
)

var (
	datosCompartidos []models.CreateEmailCMD
	wgExtract        sync.WaitGroup
	wgIndexar        sync.WaitGroup
	mutex            sync.Mutex
)

func IngestaDeDatos() (bool, error) {
	//initial variables
	numberGoroutinesWrite := 6
	numberGoroutinesRead := 6

	//instance new Zincclient and check if index exist
	client := zincClient.NewZincSearchClient()
	exists, err := client.CheckIndexExists()
	if err != nil {
		fmt.Println(err)
	}
	if exists {
		return false, nil
	}

	//create index, start index data
	errorcreateIndex := client.CreateIndex()
	if errorcreateIndex != nil {
		return false, errorcreateIndex
	}

	//read folder and return parameters
	root := "../../data/enron_mail_20110402/maildir"
	foldersArraySize, foldersArray, sizeGoroutine, error := ExtractGoRoutineConfig(root, numberGoroutinesWrite)
	if error != nil {
		return false, error
	}

	ch := make(chan models.CreateEmailCMD, numberGoroutinesWrite)

	start := time.Now()
	//define read/index goroutines
	StartIndexGoeRoutine(numberGoroutinesRead, ch, &wgIndexar, &mutex)
	StartExtractGoRoutines(numberGoroutinesWrite, foldersArraySize, foldersArray, root, ch, sizeGoroutine, &wgExtract)
	fmt.Println("entro a los wait ******************----------------------------------")
	wgExtract.Wait()
	close(ch)
	wgIndexar.Wait()
	fmt.Println("salio de los wait******************----------------------------------")
	//send last data
	error = client.StoreEmailBulk(datosCompartidos)
	if error != nil {
		fmt.Println(error)
		return true, error
	}
	duration := time.Since(start).Milliseconds()
	fmt.Printf("%dms \n", duration)
	fmt.Println("Total folders -----------------------------------: ", foldersArraySize)
	return true, nil
}
