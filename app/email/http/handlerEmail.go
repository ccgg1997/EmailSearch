package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ccgg1997/Go-ZincSearch/email/usecase"
	_ "github.com/swaggo/http-swagger"
)

type EmailHandler struct {
	emailUsecase usecase.EmailUsecase
}

func NewEmailHandler(eu usecase.EmailUsecase) *EmailHandler {
	return &EmailHandler{
		emailUsecase: eu,
	}
}

// @Summary     verify conectivity with ZincSearch
// @Description Check connectivity with ZincSearch
// @Tags        ZincSearch
// @Accept      json
// @Produce     json
// @Success     200 {string} string "La conectividad con ZincSearch esta activa, accede por medio de las peticiones HTTP de la api de email"
// @Router      /zinconection [get]
func (eh *EmailHandler) ZincSearchHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Hay peticiones en ejecucion")
	io.WriteString(w, "La conectividad con ZincSearch esta activa, accede por medio de las peticiones HTTP de la api de email")
}

// @Summary      Search text in zincsearch
// @Description  Perform a search based on the given query. Please note that the query is a string. Search results
// @Tags         Email
// @Accept       json
// @Produce      json
// @Param        query      body    QueryParam    true   "Search parameters"
// @Success       200 {string} string "Busqueda exitosa"
// @Router       /query [post]
func (eh *EmailHandler) QueryHandler(w http.ResponseWriter, r *http.Request) {

	//define and parse the body
	var requestBody map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		return
	}

	// Access to query value
	text, ok := requestBody["query"].(string)
	if !ok {
		return
	}

	//template for the zincsearch query
	var queryTemplate = `{
		"query": {
		  "bool": {
			"should": [
			  {
				"match_phrase": {
				  "content": {
					"query": "%s",
					"boost": 2
				  }
				}
			  },
			  {
				"match_phrase": {
				  "date": {
					"query": "%s",
					"boost": 1.5
				  }
				}
			  },
			  {
				"match_phrase": {
				  "xfrom": {
					"query": "%s",
					"boost": 1.6
				  }
				}
			  },
			  {
				"match_phrase": {
				  "xto": {
					"query": "%s",
					"boost": 1.6
				  }
				}
			  }
			]
		  }
		},
		"size": 40
	  }`

	//replace the %s with the text and use it as query in usecase.sentquery (IS THE HTTP REQUEST TO ZINCSEARCH)
	query := fmt.Sprintf(queryTemplate, text, text, text, text)
	email, err := eh.emailUsecase.SentQuery(query)
	if err != nil {
		http.Error(w, "Error, formato invalido del body", http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(map[string]interface{}{"EmailsEncontrados": email})
}

// QueryParam represents the structure for the search query.
// @Schema
type QueryParam struct {
	Query string `json:"query"`
}
