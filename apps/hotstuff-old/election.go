package hotstuff_old

//type SimpleElection struct {
//	addrs []string
//}
//
//func NewSimpleElection(addrs []string) *SimpleElection {
//	return &SimpleElection{addrs: addrs}
//}
//
//func (s *SimpleElection) GetLeader(round int64) string {
//	idx := (round - 1) % int64(len(s.addrs))
//	return s.addrs[idx]
//}
//
//func (s *SimpleElection) GetValidators(round int64) []string {
//	return s.addrs
//}
//
//func (s *SimpleElection) GetIntAddress(str string) string {
//	return ""
//}
