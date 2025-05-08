package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/LucianoGiope/posgolangdesafio2/pkg/httpResponseErr"
)

type Address struct {
	Cep     string `json:"cep"`
	Rua     string `json:"logradouro"`
	Bairro  string `json:"bairro"`
	Cidade  string `json:"localidade"`
	Estado  string `json:"uf"`
	Apiname string `json:"apiname"`
}

func main() {

	currencyCEP := strings.Join(os.Args[1:], " ")

	fmt.Printf("Iniciando a busca do CEP [%s]\n", currencyCEP)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/BuscaCep/"+currencyCEP, nil)
	if err != nil {
		log.Fatalf("__Falha na requisição, tente novamente !!")
	}

	res, err := http.DefaultClient.Do(req)
	if ctx.Err() != nil {
		log.Fatalf("__SERVER demorou para responder a Busca. Tente novamente!! \n %v", ctx.Err())
	}
	if err != nil {
		log.Fatalf("Erro ao chamar o SERVER\n__%v\n", err.Error())
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		var httpErrorType httpResponseErr.SHttpError
		jsonBody, err := io.ReadAll(res.Body)
		if err != nil {
			log.Fatalf("\nErro ao ler corpo da mensagem de erro\n__%v\n", err.Error())
		}

		msgErro, err := httpErrorType.DisplayMessage(jsonBody)
		if err != nil {
			log.Fatalf("Erro ao converter resposta.\n__[MESSAGE]%v\n", err.Error())
		}
		log.Fatalf("Falha durante a cotação\n%v\n", msgErro)

	} else {

		var dataResult Address
		jsonBody, err := io.ReadAll(res.Body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "\nErro ao ler a resposta Body:%v\n", err.Error())
		}

		err = json.Unmarshal(jsonBody, &dataResult)
		if err != nil {
			fmt.Fprintf(os.Stderr, "\nErro ao converter a resposta Body:%v\n", err.Error())
		}

		fmt.Printf("Você foi atendido por %s\n ENDEREÇO RETORNADO\n Rua: %s\n Bairro: %s\n Cidade: %s\n Estado: %s\n CEP: %s\n", dataResult.Apiname, dataResult.Rua, dataResult.Bairro, dataResult.Cidade, dataResult.Estado, dataResult.Cep)
	}

}
