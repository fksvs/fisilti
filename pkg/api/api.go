package api

import (
	"time"
	"errors"
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/fksvs/fisilti/pkg/storage"
	"github.com/fksvs/fisilti/pkg/cipher"
)

type API struct {
	SecretStore *storage.Storage
	MasterKey []byte
	Router *gin.Engine
}

type secret struct {
	Data string `json:"data"`
	Duration int `json:"duration"`
}

func NewAPI(secretStore *storage.Storage, masterKey []byte) *API {
	var api API

	router := gin.Default()

	api.SecretStore = secretStore
	api.MasterKey = masterKey
	api.Router = router

	router.POST("/api/v1/secret", api.createSecret)
	router.GET("/api/v1/secret/:id", api.getSecret)

	return &api
}

func (a *API) createSecret(c *gin.Context) {
	var newSecret secret

	if err := c.BindJSON(&newSecret); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to serialize request body"})
		return
	}

	ciphertext, err := cipher.EncryptData(a.MasterKey, []byte(newSecret.Data))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal encryption error"})
		return
	}

	ttl := time.Duration(newSecret.Duration) * time.Second
	id, err := a.SecretStore.CreateEntry(ciphertext, ttl)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal storage error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": string(id)})
}

func (a *API) getSecret(c *gin.Context) {
	id := c.Param("id")

	data, err := a.SecretStore.GetAndDelete(id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Secret not found"})
			return
		}
		if errors.Is(err, storage.ErrExpired) {
			c.JSON(http.StatusGone, gin.H{"error": "Secret expired"})
			return
		}
	}

	plaintext, err := cipher.DecryptData(a.MasterKey, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal decryption error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": string(plaintext)})
}
