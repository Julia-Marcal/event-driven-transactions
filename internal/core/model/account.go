package model

type Account struct {
	ID      string  `bson:"_id" json:"id"`
	Balance float64 `bson:"balance" json:"balance"`
}
