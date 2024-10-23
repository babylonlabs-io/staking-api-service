package model

import "time"

type TermsAcceptance struct {
	Address       string    `bson:"address"`
	PublicKey     string    `bson:"public_key"`
	TermsAccepted bool      `bson:"terms_accepted"`
	UpdatedAt     time.Time `bson:"updated_at"`
}
