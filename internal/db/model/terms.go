package model

import "time"

type TermsAcceptance struct {
	Id            string    `bson:"_id"`
	Address       string    `bson:"address"`
	PublicKey     string    `bson:"public_key"`
	TermsAccepted bool      `bson:"terms_accepted"`
	UpdatedAt     time.Time `bson:"updated_at"`
}
