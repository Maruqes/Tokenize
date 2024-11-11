package Tokenize

import (
	"fmt"
	"log"
	"os"

	_ "github.com/joho/godotenv/autoload"
)

//tipos de admin -> 0:superadmin, 1:admin, -1:sem acesso
//para registrar tens de ser admin 0
// tipo_admin   permissoes
// 0			//tudo porque é 0
// 1            //loja, produto

//criar conta-sistema de pagamentos

//criar conta normal  com perm -1
//form de pagamento->pagar   //dar para pagar em dinheiro
//é membro

//uma subricicao
//dar duracao a subscricao

func PrintEnv() {
	envVars := []string{"PUBLISHABLE_KEY", "SECRET_KEY", "SUBSCRIPTION_PRICE_ID"}
	for _, envVar := range envVars {
		value := os.Getenv(envVar)
		if value == "" {
			log.Panicf("%s is not set", envVar)
		} else {
			fmt.Printf("%s: %s\n\n", envVar, value)
		}
	}
}

func Init() {
	fmt.Println("Init")
	PrintEnv()
}
