package gateway

import (
	"encoding/json"
	"errors"
	"fmt"

	"os"

	zincClient "github.com/ccgg1997/Go-ZincSearch/internal/zincsearch"
)

type EmailGateway struct {
	index    string
	username string
	password string
}

func NewEmailGateway(index string) EmailGateway {
	return EmailGateway{
		index:    index,
		username: os.Getenv("ZINC_FIRST_ADMIN_USER"),
		password: os.Getenv("ZINC_FIRST_ADMIN_PASSWORD"),
	}
}

func (eg *EmailGateway) SearchQuery(query string) ([]byte, error) {

	//instance new Zincclient
	client := zincClient.NewZincSearchClient()
	responseBody, err := client.SearchDocuments(query)
	if err != nil {
		return nil, err
	}

	//extract the hits.hits part of the response and convert to json
	hitsHitsJSON, err := ExtractHits(responseBody)
	if err != nil {
		return nil, err
	}

	fmt.Println("Búsqueda realizada con éxito")
	return hitsHitsJSON, nil
}

func ExtractHits(responseBody map[string]interface{}) ([]byte, error) {

	hits, ok := responseBody["hits"].(map[string]interface{})
	if !ok {
		return nil, errors.New("no se encontró la estructura 'hits' en la respuesta")
	}

	hitsHits, ok := hits["hits"].([]interface{})
	if !ok {
		return nil, errors.New("no se encontró la estructura 'hits.hits' en la respuesta")
	}

	// Convierte la parte de "hits.hits" de nuevo a JSON
	hitsHitsJSON, err := json.Marshal(hitsHits)
	if err != nil {
		return nil, err
	}
	return hitsHitsJSON, nil
}
