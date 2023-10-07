package script

import (
	"bufio"
	"fmt"
	"io"
	"net/mail"
	"os"
	"path/filepath"
	"sync"

	zincClient "github.com/ccgg1997/Go-ZincSearch/internal/zincsearch"
	"github.com/ccgg1997/Go-ZincSearch/modules/email/models"
)

func ExtractGoRoutineConfig(root string, goroutines int) (int, []string, int, error) {

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

func StartIndexGoeRoutine(numberGoroutinesRead int, ch chan models.CreateEmailCMD, wgIndex *sync.WaitGroup, mutex *sync.Mutex) {
	//define read goroutines
	for i := 1; i <= numberGoroutinesRead; i++ {
		wgIndex.Add(1)
		go indexData(i, ch, &wgIndexar, mutex)
	}
}

func StartExtractGoRoutines(numberGoroutinesWrite int, foldersArraySize int, foldersArray []string, root string, ch chan models.CreateEmailCMD, sizeGoroutine int, wgExtract *sync.WaitGroup) { //define write goroutines
	folderstart := 0

	for i := 0; i < numberGoroutinesWrite; i++ {
		if i == numberGoroutinesWrite-1 || (foldersArraySize < numberGoroutinesWrite) {
			segment := foldersArray[folderstart:foldersArraySize]
			fmt.Println("segment: ", i, segment)
			wgExtract.Add(1)
			go SentmailtoChannel(segment, root, ch, wgExtract)
			break
		}
		//url:= string(root+foldersArray[i])
		segment := foldersArray[folderstart : folderstart+sizeGoroutine]
		wgExtract.Add(1)
		go SentmailtoChannel(segment, root, ch, wgExtract)
		fmt.Println("segment: ", i, segment)
		folderstart = folderstart + sizeGoroutine
	}

}

func SentmailtoChannel(folders []string, path string, ch chan models.CreateEmailCMD, wg *sync.WaitGroup) error {
	defer wg.Done()

	// Walk files
	for _, folder := range folders {
		root := path + "/" + folder
		fmt.Println("root:*********-*********** ", root)
		filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				wg.Add(1)
				go ExtractEmail(path, ch, wg)
			}
			return nil
		})
	}

	return nil
}

func ExtractEmail(path string, ch chan models.CreateEmailCMD, wg *sync.WaitGroup) (bool, error) {
	defer wg.Done()
	file, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	email, err := mail.ReadMessage(reader)
	if err != nil {
		return false, err
	}

	from := email.Header.Get("From")
	toAddresses, _ := email.Header.AddressList("To")
	subject := email.Header.Get("Subject")
	date := email.Header.Get("Date")
	xFrom := email.Header.Get("X-From")
	xTo := email.Header.Get("X-To")
	to := ""
	if len(toAddresses) > 0 {
		to = toAddresses[0].Address
	}
	bodyByte, err := io.ReadAll(email.Body)
	if err != nil {
		return false, err
	}
	body := string(bodyByte)
	folder := email.Header.Get("X-Folder")
	fmt.Println(folder + " " + from)
	ch <- models.CreateEmailCMD{
		Date:    date,
		From:    from,
		To:      to,
		Subject: subject,
		XFrom:   xFrom,
		XTo:     xTo,
		Content: body,
		Folder:  folder,
	}

	fmt.Println("enviando dato--------: ", folder)
	return true, nil
}

func indexData(id int, ch chan models.CreateEmailCMD, wg *sync.WaitGroup, mutex *sync.Mutex) {
	defer wg.Done()

	for dato := range ch {
		mutex.Lock()
		datosCompartidos = append(datosCompartidos, dato)
		limit := 10000
		if len(datosCompartidos) >= limit {

			client := zincClient.NewZincSearchClient()
			error := client.StoreEmailBulk(datosCompartidos)
			if error != nil {
				fmt.Println(error)
			}
			datosCompartidos = []models.CreateEmailCMD{}
		}
		mutex.Unlock()
	}
}
