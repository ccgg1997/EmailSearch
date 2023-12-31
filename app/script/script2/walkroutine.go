package script

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/mail"
	"os"
	"path/filepath"
	"sync"
	"time"

	zincClient "github.com/ccgg1997/Go-ZincSearch/internal/zincsearch"
	"github.com/ccgg1997/Go-ZincSearch/modules/email/models"
)

func IngestaDeDatos() (bool, error) {

	//instance new Zincclient
	client := zincClient.NewZincSearchClient()

	exists, err := client.CheckIndexExists()
	if err != nil {
		log.Printf("Error en la consulta del index: %v", err)
	}

	if !exists {
		//si el index no existe, se crea el index y se inicia la ingesta de datos
		start := time.Now()
		CreateIndex()
		IndexEmailData()
		duration := time.Since(start).Milliseconds()
		fmt.Printf("%dms \n", duration)
		fmt.Println("Total folders -----------------------------------: ")
		//emails := IndexEmailData()
		//fmt.Println(emails)
		return true, nil
	}

	return false, nil

}

func IndexEmailData() []models.CreateEmailCMD {
	root := "../../data/enron_mail_20110402/maildir"
	var emails []models.CreateEmailCMD
	var errorMails []string
	ch := make(chan models.CreateEmailCMD, 1)
	var wg sync.WaitGroup
	go func() {
		// Collect emails from channel
		for email := range ch {
			emails = append(emails, email)
			if len(emails) == 9000 {
				client := zincClient.NewZincSearchClient()
				client.StoreEmailBulk(emails)
				emails = []models.CreateEmailCMD{}
			}
		}
	}()
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

	// Store any remaining emails
	if len(emails) > 0 {
		fmt.Println("Almacenando los correos electrónicos restantes")
		client := zincClient.NewZincSearchClient()
		client.StoreEmailBulk(emails)
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
		Content: body,
		Folder:  folder,
	}
	return true
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
