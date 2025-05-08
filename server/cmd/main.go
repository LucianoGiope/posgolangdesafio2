package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/LucianoGiope/posgolangdesafio2/pkg/httpResponseErr"
)

type AddressViaCep struct {
	Cep    string `json:"cep"`
	Rua    string `json:"logradouro"`
	Bairro string `json:"bairro"`
	Cidade string `json:"localidade"`
	Estado string `json:"uf"`
}

type AddressBrasilApi struct {
	Cep    string `json:"cep"`
	Rua    string `json:"street"`
	Bairro string `json:"neighborhood"`
	Cidade string `json:"city"`
	Estado string `json:"state"`
}

type Address struct {
	Cep     string `json:"cep"`
	Rua     string `json:"logradouro"`
	Bairro  string `json:"bairro"`
	Cidade  string `json:"localidade"`
	Estado  string `json:"uf"`
	Apiname string `json:"apiname"`
}

func NewAddress(rua, bairro, cidade, cep, estado, apiname string) *Address {
	return &Address{cep, rua, bairro, cidade, estado, apiname}
}

func main() {
	println("\nIniciando o servidor na porta 8080 e aguardando requisições")

	routers := http.NewServeMux()

	routers.HandleFunc("/", searchCEPHandler)
	routers.HandleFunc("/BuscaCep/{cep}", searchCEPHandler)
	err := http.ListenAndServe(":8080", routers)
	if err != nil {
		log.Fatal(err)
	}

}

func searchCEPHandler(w http.ResponseWriter, r *http.Request) {

	var msgErro *httpResponseErr.SHttpError

	w.Header().Set("Content-Type", "application/json")

	urlAccess := strings.Split(r.URL.Path, "/")[1]
	if urlAccess != "BuscaCep" {
		println("The access must by in  of the endpoint http://localhost:8080/BuscaCep")
		msgErro = httpResponseErr.NewHttpError("The access must by in  of the endpoint http://localhost:8080/BuscaCep\n", http.StatusNotFound)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(msgErro)
		return
	}
	CepCurrency := r.PathValue("cep")
	if CepCurrency == "" {
		println("CEP currency not send in parameter")
		msgErro = httpResponseErr.NewHttpError("Type currency not send in parameter.\n Exemple: http://localhost:8080/BuscaCep/{CepCurrency}\n", http.StatusBadRequest)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(msgErro)
		return
	}

	regex := regexp.MustCompile(`[^0-9]+`)
	apenasNumeros := regex.ReplaceAllString(CepCurrency, "")
	if len(apenasNumeros) <= 0 {
		fmt.Println("Um CEP deve ser informado, tente novamente !!")
		msgErro = httpResponseErr.NewHttpError("Um CEP deve ser informado, tente novamente !!", http.StatusBadRequest)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(msgErro)
		return
	}
	if len(apenasNumeros) != 8 {
		fmt.Printf("O CEP %s não é um número válido, tente novamente !!", CepCurrency)
		msgErro = httpResponseErr.NewHttpError(fmt.Sprintf("O CEP %s não é um número válido, tente novamente !!", CepCurrency), http.StatusBadRequest)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(msgErro)
		return
	}

	ctxClient := r.Context()

	timeAtual := time.Now()
	fmt.Printf("\n-> Searching ADDRESS for the ZIPCODE:%s in %v.\n", CepCurrency, timeAtual.Format("02/01/2006 15:04:05 ")+timeAtual.String()[20:29]+" ms")

	chanelViaCep := make(chan bool)
	chanelBrasilApi := make(chan bool)

	errCode := 0
	errText := ""
	// var dataAdressResult Address

	ctxSearchViaCep, cancelSearchVC := context.WithTimeout(ctxClient, time.Second*1)
	defer cancelSearchVC()
	urlSearch := fmt.Sprintf("https://viacep.com.br/ws/%s/json/", CepCurrency)

	go func() {
		resBody, err := searchCep(ctxSearchViaCep, urlSearch, "ViaCep")
		if err != nil {
			msgErrFix := "__Error searching for ViaCep."
			if ctxSearchViaCep.Err() != nil {
				errCode = http.StatusRequestTimeout
				errText = msgErrFix + "\n____[MESSAGE] Tempo de pesquisa excedido"
			} else {
				errCode = http.StatusBadRequest
				errText = msgErrFix + "\n____[MESSAGE] Falha na requisição."
			}

			fmt.Printf("Voltei da busca de CEP ViaCep no erro:\n")
			msgErro = httpResponseErr.NewHttpError(errText, errCode)
			w.WriteHeader(errCode)
			json.NewEncoder(w).Encode(msgErro)
		}
		if resBody != nil {
			var viaCepResult AddressViaCep
			err = json.Unmarshal(resBody, &viaCepResult)
			if err != nil {
				fmt.Fprintf(os.Stderr, "\nErro ao converter a resposta resBody:%v\n", err.Error())
			}
			dataAdressResult := NewAddress(viaCepResult.Rua, viaCepResult.Bairro, viaCepResult.Cidade, viaCepResult.Cep, viaCepResult.Estado, "ViaCep")

			fmt.Printf("Voltei da busca de CEP ViaCep resBody:\n")
			fmt.Println(viaCepResult)
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(dataAdressResult)
		}
		chanelViaCep <- true
	}()

	ctxSearchBrasilApi, cancelSearchBA := context.WithTimeout(ctxClient, time.Second*1)
	defer cancelSearchBA()

	go func() {
		urlSearch := fmt.Sprintf("https://brasilapi.com.br/api/cep/v1/%s", CepCurrency)
		resBody, err := searchCep(ctxSearchBrasilApi, urlSearch, "BrasilApi")
		if err != nil {
			msgErrFix := "__Error searching for BrasilApi."

			if ctxSearchBrasilApi.Err() != nil {
				errCode = http.StatusRequestTimeout
				errText = msgErrFix + "\n____[MESSAGE] Tempo de pesquisa excedido"
			} else {
				errCode = http.StatusBadRequest
				errText = msgErrFix + "\n____[MESSAGE] Falha na requisição."
			}
			fmt.Printf("Voltei da busca de CEP BrasilApi no erro:\n")
			msgErro = httpResponseErr.NewHttpError(errText, errCode)
			w.WriteHeader(errCode)
			json.NewEncoder(w).Encode(msgErro)
		}
		if resBody != nil {
			var brasilApiResult AddressBrasilApi
			err = json.Unmarshal(resBody, &brasilApiResult)
			if err != nil {
				fmt.Fprintf(os.Stderr, "\nErro ao converter a resposta resBody:%v\n", err.Error())
			}

			dataAdressResult := NewAddress(brasilApiResult.Rua, brasilApiResult.Bairro, brasilApiResult.Cidade, brasilApiResult.Cep, brasilApiResult.Estado, "BrasilApi")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(dataAdressResult)
		}
		chanelBrasilApi <- true
	}()

	select {
	case <-chanelViaCep:
		cancelSearchBA()
		cancelSearchVC()
		<-chanelBrasilApi

	case <-chanelBrasilApi:
		cancelSearchVC()
		cancelSearchBA()
		<-chanelViaCep
	}

	fmt.Printf("\n-> Time total in milliseconds traveled %v.\n\n", time.Since(timeAtual))
}

func searchCep(ctx context.Context, urlSearch string, apiName string) ([]byte, error) {

	timeAtual := time.Now()
	fmt.Printf("\n--> Iniciando busca tipo:%s em %s\n", apiName, timeAtual.Format("02/01/2006 15:04:05 ")+timeAtual.String()[20:29]+" ms")

	bodyResp, err := requestCep(ctx, urlSearch)

	select {
	case <-ctx.Done():
		err2 := ctx.Err()
		if err2 == context.Canceled {
			fmt.Printf("\n________Cancelada a consultar com fornecedor %s\n", apiName)
		} else if err2 == context.DeadlineExceeded {
			timeAtual = time.Now()
			fmt.Printf("\n________Tempo excedido para consultar com fornecedor %s em %s\n", apiName, timeAtual.Format("02/01/2006 15:04:05 ")+timeAtual.String()[20:29]+" ms")
		} else {
			fmt.Printf("\n________Abandonada a consulta em %s por motivo desconhecido.\n [ERROR] %v\n", apiName, err)
		}
		return nil, nil
	default:
		if err != nil {
			fmt.Printf("\n________Falha ao consultar CEP em %s\n [MESSAGE]%v\n", apiName, err.Error())
			return nil, err
		} else {
			timeAtual = time.Now()
			fmt.Printf("\n________Capturados os dados do CEP com API: %s em %s\n", apiName, timeAtual.Format("02/01/2006 15:04:05 ")+timeAtual.String()[20:29]+" ms")

			return *bodyResp, nil
		}
	}
}

func requestCep(ctx context.Context, urlSearch string) (*[]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", urlSearch, nil)
	if err != nil {
		return nil, err
	}

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	bodyResp, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	return &bodyResp, nil
}
