## Primeiro desafio proposto pelo curso.

## Texto descritivo do desafio
 Neste desafio você terá que usar o que aprendemos com Multithreading e APIs para buscar o resultado mais rápido entre duas APIs distintas.

As duas requisições serão feitas simultaneamente para as seguintes APIs:

https://brasilapi.com.br/api/cep/v1/01153000 + cep

http://viacep.com.br/ws/" + cep + "/json/

Os requisitos para este desafio são:

- Acatar a API que entregar a resposta mais rápida e descartar a resposta mais lenta.

- O resultado da request deverá ser exibido no command line com os dados do endereço, bem como qual API a enviou.

- Limitar o tempo de resposta em 1 segundo. Caso contrário, o erro de timeout deve ser exibido.
 
Ao finalizar, envie o link do repositório para correção.

### Como utilizar a API

## Primeiro deve-se subir o server
- Acesse a pasta server/cmd
- Execute o comando 
- - go run main.go

## Com servidor ativado, executar a consulta
- Utilizando o cliente em Go
- - Acesse a pasta client/cmd e rode o comando abaixo
- - go run main.go {NÚMERO DO CEP}



