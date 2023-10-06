package script

import (
	"os"
	"fmt"

	"github.com/ccgg1997/Go-ZincSearch/modules/email/models"
)

func WriteGoRoutineConfig(root string, goroutines int) (int, []string, int, error) {

	//open folder
	folder, error := os.Open(root)
	if error != nil {
		return 0, nil, 0, error
	}
	defer folder.Close()

	//read folder names
	foldersName, err := folder.Readdirnames(0)
	if err != nil {
		return 0, nil, 0, err
	}

	//calculate size of each goroutine
	foldersNameSize := len(foldersName)
	sizeGoroutine := foldersNameSize / goroutines
	return foldersNameSize, foldersName, sizeGoroutine, nil
}

func StartReadGoeRoutine(numberGoroutinesRead int , ch chan models.CreateEmailCMD )(){	
	//define read goroutines
	for i := 1; i <= numberGoroutinesRead; i++ {
		wgIndexar.Add(1)
		go indexarDatos(i, ch, &wgIndexar)
	}
}

func StartIndexGoRoutines(numberGoroutinesWrite int, foldersArraySize int, foldersArray []string,root string, ch chan models.CreateEmailCMD, sizeGoroutine int )(){//define write goroutines
	folderstart := 0
	for i := 0; i < numberGoroutinesWrite; i++ {
		if i == numberGoroutinesWrite-1 || (foldersArraySize < numberGoroutinesWrite) {
			segment := foldersArray[folderstart:foldersArraySize]
			fmt.Println("segment: ", i, segment)
			wgEscritura.Add(1)
			go SentmailtoChannel(segment, root, ch, &wgEscritura)
			break
		}
		//url:= string(root+foldersArray[i])
		segment := foldersArray[folderstart : folderstart+sizeGoroutine]
		wgEscritura.Add(1)
		go SentmailtoChannel(segment, root, ch, &wgEscritura)
		fmt.Println("segment: ", i, segment)
		folderstart = folderstart + sizeGoroutine
	}
	
}