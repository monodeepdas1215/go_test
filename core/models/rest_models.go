package models


type AddTransactionRequestJson struct {
	Amount		float64	`json:"amount"`
	Type		string	`json:"type"`
	ParentId	int64	`json:"parent_id"`
}