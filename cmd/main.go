package main

import (
	"log"
	"github.com/fksvs/fisilti/pkg/api"
	"github.com/fksvs/fisilti/pkg/storage"
        "github.com/fksvs/fisilti/pkg/cipher"
)

func main() {
	secretStore := storage.InitStorage()
	masterKey, err := cipher.GenerateMasterKey(32)
	if err != nil {
		log.Panicf("failed to generate master key: %v", err)
	}

	api := api.NewAPI(secretStore, masterKey)

	api.Router.RunTLS(":8080", "./certs/cert.pem", "./certs/key.pem")
}
