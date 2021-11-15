package hotstuff

type SimpleElection struct {
	addrs []string
}

func NewSimpleElection(addrs []string) *SimpleElection {
	return &SimpleElection{addrs: addrs}
}

func (s *SimpleElection) GetLeader(round int64) string {
	pos := round % 3
	return s.addrs[pos]
}

func (s *SimpleElection) GetValidators(round int64) []string {
	return s.addrs
}

func (s *SimpleElection) GetIntAddress(str string) string {
	return ""
}
