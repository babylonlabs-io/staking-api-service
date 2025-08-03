package service

// AllowListService provides in-memory lookup for staking transaction hashes
// that are eligible for canExpand evaluation. The allow-list is loaded once
// during API initialization and never changes during runtime.
type AllowListService struct {
	allowList map[string]bool
}

func NewAllowListService() (*AllowListService, error) {
	return &AllowListService{
		allowList: make(map[string]bool),
	}, nil
}

// LoadAllowList loads staking transaction hashes into memory for runtime lookup.
// This is called once during API initialization and never changes afterward.
func (a *AllowListService) LoadAllowList(stakingHashes []string) error {
	// Initialize the map with the correct capacity
	a.allowList = make(map[string]bool, len(stakingHashes))

	for _, hash := range stakingHashes {
		a.allowList[hash] = true
	}

	return nil
}

// IsInAllowList checks if a staking transaction hash exists in the allow-list.
// This method provides O(1) lookup performance for runtime canExpand evaluation.
// Since the map is read-only after initialization, no locking is needed.
func (a *AllowListService) IsInAllowList(stakingTxHashHex string) bool {
	return a.allowList[stakingTxHashHex]
}

// GetAllowListSize returns the number of entries in the allow-list
func (a *AllowListService) GetAllowListSize() int {
	return len(a.allowList)
}
