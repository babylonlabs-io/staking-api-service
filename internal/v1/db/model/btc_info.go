package v1dbmodel

const LatestBtcInfoId = "latest"

type BtcInfo struct {
	ID             string `bson:"_id"`
	BtcHeight      uint64 `bson:"btc_height"`
	ConfirmedTvl   uint64 `bson:"confirmed_tvl"`
	UnconfirmedTvl uint64 `bson:"unconfirmed_tvl"`
}
