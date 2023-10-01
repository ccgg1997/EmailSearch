package zincsearch

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	_ "net/http/pprof"
)

type ZincSearchClient struct {
	Url      string
	User     string
	Password string
}

func NewZincSearchClient(url string, user string, password string) *ZincSearchClient {
	return &ZincSearchClient{
		Url:	  url,
		User:     user,
		Password: password,
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

func (n *ZincSearchClient) SearchDocuments(url string, query string) (map[string]interface{},error ){

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
		return nil, errors.New("error al realizar la búsqueda, estado de la petición: " + resp.Status)
	}

	//convert the response to map
	var responseBody map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		return nil, err
	}
	return responseBody, nil
}

