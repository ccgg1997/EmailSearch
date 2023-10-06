package script

import (
	"fmt"
	"sync"

	zincClient "github.com/ccgg1997/Go-ZincSearch/internal/zincsearch"
	"github.com/ccgg1997/Go-ZincSearch/modules/email/models"
)

var (
	datosCompartidos []models.CreateEmailCMD
	datosCompartidosAux []models.CreateEmailCMD
	wgEscritura      sync.WaitGroup
	wgIndexar        sync.WaitGroup
	mu                  sync.Mutex

)

func IngestaDeDatos() (bool, error) {
	//initial variables
	numberGoroutinesWrite := 8
	numberGoroutinesRead := 1

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
	root := "../../data/enron_mail_20110402/maildir/allen-p"
	foldersArraySize, foldersArray, sizeGoroutine, error := WriteGoRoutineConfig(root, numberGoroutinesWrite)
	if error != nil {
		return false, error
	}

	ch := make(chan models.CreateEmailCMD, numberGoroutinesWrite)
	
	//define read/write goroutines
	StartReadGoeRoutine(numberGoroutinesRead, ch)
	StartIndexGoRoutines(numberGoroutinesWrite, foldersArraySize, foldersArray, root, ch, sizeGoroutine)
	wgEscritura.Wait()
	close(ch)
	wgIndexar.Wait()

	//send last data
	error = client.StoreEmailBulk(datosCompartidos)
	if error != nil {
		fmt.Println(error)
	}
	fmt.Println("Total folders -----------------------------------: ", foldersArraySize)
	return true, nil
}



