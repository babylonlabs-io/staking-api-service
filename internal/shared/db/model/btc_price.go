package dbmodel

import "time"

const SymbolBTC = "btc"

type CoinPrice struct {
	ID        string    `bson:"_id"` // symbol
	Price     float64   `bson:"price"`
	CreatedAt time.Time `bson:"created_at"` // TTL index will be on this field
}
