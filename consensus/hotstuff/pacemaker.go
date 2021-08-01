// Copyright Xuperchain Authors
// link: https://github.com/xuperchain/xupercore

package hotstuff

// IPacemaker is the interface of Pacemaker. It responsible for generating a new round.
// We assume Pacemaker in all correct replicas will have synchronized leadership after GST.
// Safty is entirely decoupled from liveness by any potential instantiation of Packmaker.
// Different consensus have different pacemaker implement
type IPacemaker interface {
	// CurrentView return current view of this node.
	GetCurrentView() int64
	// 原NextNewProposal，generate new proposal directly.
	AdvanceView(qc IQuorumCert) (bool, error)
}

// DefaultPaceMaker 是一个IPacemaker的默认实现，我们与IPacemaker放置在一起，方便查看
// IPacemaker的新实现直接直接替代DefaultPaceMaker即可
// The Pacemaker keeps track of votes and of time.
// TODO:  the Pacemaker broadcasts a TimeoutMsg notification.
type DefaultPaceMaker struct {
	CurrentView int64
	// timeout int64
}

func (p *DefaultPaceMaker) AdvanceView(qc IQuorumCert) (bool, error) {
	r := qc.GetProposalView()
	if r+1 > p.CurrentView {
		p.CurrentView = r + 1
	}
	return true, nil
}

func (p *DefaultPaceMaker) GetCurrentView() int64 {
	return p.CurrentView
}
