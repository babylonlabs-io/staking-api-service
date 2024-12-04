package model

import "time"

const BtcPriceDocID = "btc_price"

type BtcPrice struct {
	ID        string    `bson:"_id"` // primary key, will always be "btc_price" to ensure single document
	Price     float64   `bson:"price"`
	CreatedAt time.Time `bson:"created_at"` // TTL index will be on this field
}
