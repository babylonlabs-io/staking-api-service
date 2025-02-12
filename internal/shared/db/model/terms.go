package dbmodel

import "go.mongodb.org/mongo-driver/bson/primitive"

type TermsAcceptance struct {
	Id        primitive.ObjectID `bson:"_id,omitempty"`
	Address   string             `bson:"address"`
	PublicKey string             `bson:"public_key"`
}
