package data_store

import "github.com/monodeepdas1215/go_test/core/models"

var AppDb AppDataStore

type AppDataStore interface {
	Connect() error
	Disconnect()
	UpsertTransactionDetails(details map[string]interface{}) error
	GetTransactionDetails(query map[string]interface{}) ([]models.Transactions, error)
}
