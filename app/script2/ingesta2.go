package scripts

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/mail"
	"os"
	"path/filepath"
	"sync"

	"github.com/ccgg1997/Go-ZincSearch/modules/email/models"
	zincClient "github.com/ccgg1997/Go-ZincSearch/internal/zincsearch"
)

func IngestaDeDatos() bool {
	
	//instance new Zincclient
	client := zincClient.NewZincSearchClient()

	exists, err := client.CheckIndexExists()
	if err != nil {
		log.Printf("Error en la consulta del index: %v", err)
	}

	if !exists {
		//si el index no existe, se crea el index y se inicia la ingesta de datos
		CreateIndex()
		readEmailData()
		//emails := readEmailData()
		//fmt.Println(emails)
		return true
	}

	return false

}


func readEmailData() []models.CreateEmailCMD {
	root := "../../data/enron_mail_20110402/maildir/allen-p"
	var emails []models.CreateEmailCMD
	var errorMails []string
	ch := make(chan models.CreateEmailCMD, 14)
	var wg sync.WaitGroup

	// Walk files
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			errorMails = append(errorMails, fmt.Sprintf("Error al acceder al archivo: %s", err))
			return nil
		}
		if !info.IsDir() {
			wg.Add(1)
			go processFile(path, ch, &wg, &errorMails)
		}
		return nil
	})

	// Close channel after all goroutines finish
	go func() {
		wg.Wait()
		close(ch)
	}()

	// Collect emails from channel
	for email := range ch {
		emails = append(emails, email)
		if len(emails) == 9000 {
			storeEmail(emails)
			emails = []models.CreateEmailCMD{}
		}
	}

	// Store any remaining emails
	if len(emails) > 0 {
		fmt.Println("Almacenando los correos electrónicos restantes")
		storeEmail(emails)
	}

	// Aquí puedes manejar o imprimir los correos electrónicos erróneos si lo deseas
	for _, errorMsg := range errorMails {
		fmt.Println(errorMsg)
	}

	return emails
}

func processFile(path string, ch chan models.CreateEmailCMD, wg *sync.WaitGroup, errorMails *[]string) bool {
	defer wg.Done()

	file, err := os.Open(path)
	if err != nil {
		*errorMails = append(*errorMails, fmt.Sprintf("Error al abrir el archivo: %s", err))
		return false
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	email, err := mail.ReadMessage(reader)
	if err != nil {
		*errorMails = append(*errorMails, fmt.Sprintf("Error al leer el correo electrónico: %s, path: %s", err, path))
		return false
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
		*errorMails = append(*errorMails, fmt.Sprintf("Error al leer el cuerpo del correo electrónico: %s, path: %s", err, path))
		return false
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
		Content:    body,
		Folder:  folder,
	}
	return true
}

func storeEmail(emails []models.CreateEmailCMD) error {
	payload := struct {
		Index   string                  `json:"index"`
		Records []models.CreateEmailCMD `json:"records"`
	}{
		Index:   "email",
		Records: emails,
	}

	emailJSON, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}

	url := os.Getenv("ZINC_API_URL") + "/api/" + "/_bulkv2"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(emailJSON))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("accept", "application/json")
	req.SetBasicAuth(os.Getenv("ZINC_FIRST_ADMIN_USER"), os.Getenv("ZINC_FIRST_ADMIN_PASSWORD"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("error al almacenar el email en ZincSearch")
	}

	return nil
}



func CreateIndex() error {

	texto := `{
		"name": "email",
		"settings": {
			"number_of_shards": 3,
			"number_of_replicas": 1,
			"analysis": {
				"analyzer": {
					"correo_analyzer": {
						"type": "custom",
						"tokenizer": "standard",
						"filter": ["lowercase", "stopwords_marker"],
						"char_filter": ["html_strip"]
					}
				},
				"filter": {
					"stopwords_marker": {
						"type": "keyword_marker",
						"keywords": ["a", "an", "and", "are", "as", "at", "be", "but", "by", "for", "if", "in", "into", "is", "it", "no", "not", "of", "on", "or", "such", "that", "the", "their", "then", "there", "these", "they", "this", "to", "was", "will", "with"]
					}
				}
			}
		}
	}
	`

	// Define the request URL
	url := os.Getenv("ZINC_API_URL") + "/api/index"

	// Create a new HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(texto)))
	if err != nil {
		return err
	}

	// Set the request headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("accept", "application/json")
	req.SetBasicAuth(os.Getenv("ZINC_FIRST_ADMIN_USER"), os.Getenv("ZINC_FIRST_ADMIN_PASSWORD"))

	// Send the request using an HTTP client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		return errors.New("error al crear el índice en ZincSearch")
	}

	return nil
}
