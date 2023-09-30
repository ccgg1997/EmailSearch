package usecase

import (
	"encoding/json"
	"fmt"

	"github.com/ccgg1997/Go-ZincSearch/modules/email/gateway"
	"github.com/ccgg1997/Go-ZincSearch/modules/email/models"
)

type EmailUsecase struct {
	emailGateway gateway.EmailGateway
}

// Struct for the query
type QueryJSONData struct {
	Data []struct {
		EmailData models.CreateEmailCMD `json:"_source"`
	} `json:"data"`
}

func NewEmailUsecase(eg gateway.EmailGateway) EmailUsecase {
	return EmailUsecase{
		emailGateway: eg,
	}
}

func (eu *EmailUsecase) FindQuery(query string) ([]models.CreateEmailCMD, error) {

	// make the query to zincsearch and store the response in the struct through the gateway
	response, err := eu.emailGateway.SearchQuery(query)
	if err != nil {
		return nil, err
	}

	// Unmarshal the response into the struct(source)
	var EstuctEmailsFound QueryJSONData
	if err := json.Unmarshal([]byte(response), &EstuctEmailsFound.Data); err != nil {
		fmt.Println("Error al analizar el JSON:", err)
		return nil, err
	}

	//fill the slice with the unique emails
	emailsFound, _ := StructUniqueEmail(EstuctEmailsFound)

	// print the emails found
	fmt.Println(emailsFound)
	return emailsFound, nil
}

func StructUniqueEmail( EstuctEmailsFound QueryJSONData)([]models.CreateEmailCMD,error) {

	//instance the set and the slice
	contentSet := make(map[string]struct{})
	var uniqueEmails []models.CreateEmailCMD

	//fill the slice with the unique emails
	for _, item := range EstuctEmailsFound.Data {
		email := item.EmailData
		content := item.EmailData.Content
		// Verify if the content already exists in the set
		_, exists := contentSet[content]
		if exists {
			continue
		}
		//Add new content to the set and append the email to the slice
		contentSet[content] = struct{}{}
		uniqueEmails = append(uniqueEmails, email)
	}

	return uniqueEmails, nil
}
