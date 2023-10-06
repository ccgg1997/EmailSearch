package zincsearch

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/ccgg1997/Go-ZincSearch/modules/email/models"
	
)

type ZincSearchClient struct {
	Url      string
	User     string
	Password string
}

func NewZincSearchClient() *ZincSearchClient {
	return &ZincSearchClient{
		Url:      os.Getenv("ZINC_API_URL"),
		User:     os.Getenv("ZINC_FIRST_ADMIN_USER"),
		Password: os.Getenv("ZINC_FIRST_ADMIN_PASSWORD"),
	}
}

func (n *ZincSearchClient) CheckClient() error {
	req, err := http.Get(n.Url)
	if err != nil {
		fmt.Printf("error al conectarse a %s: %v", n.Url, err)
		return err // Retorna el error
	}
	defer req.Body.Close()
	fmt.Println(req.StatusCode)
	return nil

}

func (n *ZincSearchClient) SearchDocuments(query string) (map[string]interface{}, error) {
	//search the query
	url := n.Url + "/es/" + "email" + "/_search"
	return n.ZincRequestSearch( url, query)
}

func (n *ZincSearchClient) StoreEmailBulk(emails []models.CreateEmailCMD)error{
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
	url := n.Url + "/api/" + "/_bulkv2"
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

func (n *ZincSearchClient) ZincRequestSearch( url string, query string) (map[string]interface{}, error) {

	//set the request, header and auth
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(query)))
	if err != nil {
		fmt.Println("error creando la solicitud HTTP")
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("accept", "application/json")
	req.SetBasicAuth(n.User, n.Password)

	//do the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("error en la peticion")
		return nil, errors.New("Error al realizar la búsqueda, error en la petición" + err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("error en la respuesta de la petición")
		fmt.Println(n.User)
		fmt.Println(n.Password)
		fmt.Println(n.Url)
		return nil, errors.New("error al realizar la búsqueda, estado de la petición: " + resp.Status)
	}

	//convert the response to map
	var responseBody map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		return nil, err
	}
	return responseBody, nil
}

func (n *ZincSearchClient) CheckIndexExists() (bool, error) {
	//crear peticion
	url := n.Url + "/api/index/email/"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}

	//datos de la peticion
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("accept", "application/json")
	req.SetBasicAuth(n.User, n.Password)

	//enviar peticion
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}

	//evaluar respuesta
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false, errors.New("error: no existe el index")
	}
	return true, nil
}

func (n *ZincSearchClient)  CreateIndex() error {

	texto := `{
		"name": "email",
		"settings": {
			"number_of_shards": 1,
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