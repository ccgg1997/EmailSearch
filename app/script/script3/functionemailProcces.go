package script

import(	
	"bufio"
	"io"
	"net/mail"
	"os"
	"path/filepath"
	"sync"
	"fmt"

	zincClient "github.com/ccgg1997/Go-ZincSearch/internal/zincsearch"
	"github.com/ccgg1997/Go-ZincSearch/modules/email/models"
)

func SentmailtoChannel(folders []string, path string, ch chan models.CreateEmailCMD, wg *sync.WaitGroup) error {
	// Walk files
	for _, folder := range folders {
		root := path + "/" + folder
		fmt.Println("root:*********-*********** ", root)
		filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				ExtractEmail(path, ch, wg)
			}
			return nil
		})
	}
	wg.Done()
	return nil
}

func ExtractEmail(path string, ch chan models.CreateEmailCMD, wg *sync.WaitGroup) (bool, error) {

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

func indexarDatos(id int, ch chan models.CreateEmailCMD, wg *sync.WaitGroup) {
	for dato := range ch {
		//fmt.Printf("Goroutine %d recibió mensaje de %s\n", id, dato.Folder)
		datosCompartidos = append(datosCompartidos, dato)
		if len(datosCompartidos) == 9000 {
			client := zincClient.NewZincSearchClient()
			error := client.StoreEmailBulk(datosCompartidos)
			if error != nil {
				fmt.Println(error)
			}
			datosCompartidos = []models.CreateEmailCMD{}
		}
	}
	wg.Done()
}
