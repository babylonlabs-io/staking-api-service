package v2service

func (s *V2Service) AssessAddress(address string) (any, error) {
	return s.Clients.ChainAnalysis.AssessAddress(address)
}
