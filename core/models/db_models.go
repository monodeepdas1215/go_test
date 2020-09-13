package models

type Transactions struct {

	Id			int64	`gorm:"index:id" json:"id"`
	ParentId	int64	`gorm:"index:parent_id" json:"parent_id"`
	Amount 		float64	`gorm:"not null" json:"amount"`
	Type		string	`gorm:"index:type;" json:"type"`
	TotalAmount	float64	`json:"total_amount"`
}
