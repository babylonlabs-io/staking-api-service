// temp implementation of events
// TODO: add events to github.com/babylonlabs-io/staking-queue-client
package v2queueschema

const (
	ActiveStakingQueueName    string = "active_staking_queue"
	UnbondingStakingQueueName string = "unbonding_staking_queue"
)

const (
	ActiveStakingEventType    EventType = 1
	UnbondingStakingEventType EventType = 2
)

// Event schema versions, only increment when the schema changes
const (
	ActiveEventVersion    int = 0
	UnbondingEventVersion int = 0
)

type EventType int

type EventMessage interface {
	GetEventType() EventType
	GetStakingTxHashHex() string
}

type ActiveStakingEvent struct {
	SchemaVersion             int       `json:"schema_version"`
	EventType                 EventType `json:"event_type"` // always 1. ActiveStakingEventType
	StakingTxHashHex          string    `json:"staking_tx_hash_hex"`
	StakerBtcPkHex            string    `json:"staker_btc_pk_hex"`
	FinalityProviderBtcPksHex []string  `json:"finality_provider_btc_pks_hex"`
	StakingAmount             uint64    `json:"staking_amount"`
}

func (e ActiveStakingEvent) GetEventType() EventType {
	return ActiveStakingEventType
}

func (e ActiveStakingEvent) GetStakingTxHashHex() string {
	return e.StakingTxHashHex
}

func NewActiveStakingEvent(
	stakingTxHashHex string,
	stakerBtcPkHex string,
	finalityProviderBtcPksHex []string,
	stakingValue uint64,
) ActiveStakingEvent {
	return ActiveStakingEvent{
		SchemaVersion:             ActiveEventVersion,
		EventType:                 ActiveStakingEventType,
		StakingTxHashHex:          stakingTxHashHex,
		StakerBtcPkHex:            stakerBtcPkHex,
		FinalityProviderBtcPksHex: finalityProviderBtcPksHex,
		StakingAmount:             stakingValue,
	}
}

type UnbondingStakingEvent struct {
	SchemaVersion           int       `json:"schema_version"`
	EventType               EventType `json:"event_type"` // always 2. UnbondingStakingEventType
	StakingTxHashHex        string    `json:"staking_tx_hash_hex"`
	UnbondingStartHeight    uint64    `json:"unbonding_start_height"`
	UnbondingStartTimestamp int64     `json:"unbonding_start_timestamp"`
	UnbondingTimeLock       uint64    `json:"unbonding_timelock"`
	UnbondingOutputIndex    uint64    `json:"unbonding_output_index"`
	UnbondingTxHex          string    `json:"unbonding_tx_hex"`
	UnbondingTxHashHex      string    `json:"unbonding_tx_hash_hex"`
}

func (e UnbondingStakingEvent) GetEventType() EventType {
	return UnbondingStakingEventType
}

func (e UnbondingStakingEvent) GetStakingTxHashHex() string {
	return e.StakingTxHashHex
}

func NewUnbondingStakingEvent(
	stakingTxHashHex string,
	unbondingStartHeight uint64,
	unbondingStartTimestamp int64,
	unbondingTimeLock uint64,
	unbondingOutputIndex uint64,
	unbondingTxHex string,
	unbondingTxHashHex string,
) UnbondingStakingEvent {
	return UnbondingStakingEvent{
		SchemaVersion:           UnbondingEventVersion,
		EventType:               UnbondingStakingEventType,
		StakingTxHashHex:        stakingTxHashHex,
		UnbondingStartHeight:    unbondingStartHeight,
		UnbondingStartTimestamp: unbondingStartTimestamp,
		UnbondingTimeLock:       unbondingTimeLock,
		UnbondingOutputIndex:    unbondingOutputIndex,
		UnbondingTxHex:          unbondingTxHex,
		UnbondingTxHashHex:      unbondingTxHashHex,
	}
}
