package v2dbmodel

import "time"

type FinalityProviderLogo struct {
	Id        string    `bson:"_id"`
	URL       *string   `bson:"url"`
	CreatedAt time.Time `bson:"created_at"`
}
