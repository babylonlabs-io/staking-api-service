package model

import "time"

type BtcPrice struct {
	Price     float64   `bson:"price"`
	CreatedAt time.Time `bson:"created_at"` // TTL index will be on this field
}
