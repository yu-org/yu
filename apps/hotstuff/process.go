package hotstuff

import (
	"encoding/json"
	"github.com/golang/protobuf/proto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/sirupsen/logrus"
	"github.com/xuperchain/xupercore/kernel/consensus/base/driver/chained-bft/crypto"
	chainedBftPb "github.com/xuperchain/xupercore/kernel/consensus/base/driver/chained-bft/pb"
	"github.com/xuperchain/xupercore/lib/utils"
	ch "github.com/yu-org/yu/consensus/chained-hotstuff"
)

const (
	ProposeCode int = 102
	VoteCode    int = 103
)

func (h *Hotstuff) handleRecvProposal(data []byte) (noResp []byte, err error) {
	newProposal := &chainedBftPb.ProposalMsg{}
	err = proto.Unmarshal(data, newProposal)
	if err != nil {
		logrus.Errorf("smr::handleReceivedProposal Encode ProposalMsg error: %v", err)
		return
	}
	parentQC := &ch.QuorumCert{}
	err = json.Unmarshal(newProposal.GetJustifyQC(), parentQC)
	if err != nil {
		logrus.Errorf("smr::voteProposal::vote Encode parentQC error: %v", err)
		return
	}

	newVote := &ch.VoteInfo{
		ProposalId:   newProposal.GetProposalId(),
		ProposalView: newProposal.GetProposalView(),
		ParentId:     parentQC.GetProposalId(),
		ParentView:   parentQC.GetProposalView(),
	}

	needSendMsg, err := h.smr.CheckViewAndRound(newProposal, newVote, parentQC)
	if err != nil {
		return
	}
	if needSendMsg {
		leader := newProposal.GetSign().GetAddress()
		leaderPeerID, err := peer.Decode(leader)
		if err != nil {
			logrus.Errorf("smr::handleReceivedProposal Decode P2P-ID(%s) error: %v, vote view number(%d)", leader, err, newVote.ProposalView)
			return
		}
		if h.env.P2pNetwork.LocalID() != leaderPeerID {
			go h.env.P2pNetwork.RequestPeer(leaderPeerID, ProposeCode, data)
		}

	}
	voteMsg, voteTo, err := h.smr.HandleRecvProposal(newProposal, newVote, parentQC)
	if err != nil {
		return
	}
	voteMsgByt, err := proto.Marshal(voteMsg)
	if err != nil {
		logrus.Errorf("smr::voteProposal::vote  Encode VoteMsg error: %v, vote view number(%d)", err, newVote.ProposalView)
		return
	}
	votePeerID, err := peer.Decode(voteTo)
	if err != nil {
		logrus.Errorf("smr::voteProposal::vote Decode P2P-ID(%s) error: %v, vote view number(%d)", voteTo, err, newVote.ProposalView)
		return
	}

	go func() {
		_, err := h.env.P2pNetwork.RequestPeer(votePeerID, VoteCode, voteMsgByt)
		if err != nil {
			logrus.Errorf("smr::voteProposal vote to next leader(%s) error: %v,  vote view number(%d)", voteTo, err, newVote.ProposalView)
		} else {
			logrus.Debug("smr::voteProposal::vote  vote to next leader(%s)  vote view number(%d)", voteTo, newVote.ProposalView)
		}
	}()

	return
}

func (h *Hotstuff) handleRecvVoteMsg(data []byte) (noResp []byte, err error) {
	newVoteMsg := &chainedBftPb.VoteMsg{}
	err = proto.Unmarshal(data, newVoteMsg)
	if err != nil {
		logrus.Errorf("smr::handleRecvVoteMsg Encode VoteMsg error: %v", err)
		return
	}
	err = h.smr.HandleRecvVoteMsg(newVoteMsg)
	if err != nil {
		logrus.Errorf("smr::handleRecvVoteMsg handle VoteMsg error: %v", err)
	}
	return
}

func (h *Hotstuff) signProposal(msg *chainedBftPb.ProposalMsg) (*chainedBftPb.ProposalMsg, error) {
	msgDigest, err := crypto.MakeProposalMsgDigest(msg)
	if err != nil {
		return nil, err
	}
	msg.MsgDigest = msgDigest
	sig, err := h.myPrivKey.SignData(msgDigest)
	if err != nil {
		return nil, err
	}
	msg.Sign = &chainedBftPb.QuorumCertSign{
		Address:   h.LocalAddress(),
		PublicKey: h.myPubkey.StringWithType(),
		Sign:      sig,
	}
	return msg, nil
}

func (h *Hotstuff) doPropose(viewNumber int64, proposalID []byte) {
	proposal, err := h.smr.DoPropose(viewNumber, proposalID, h.ValidatorsP2pID())
	if err != nil {
		return
	}
	propMsg, err := h.signProposal(proposal)
	if err != nil {
		logrus.Error("smr::ProcessProposal SignProposalMsg error: ", err)
		return
	}
	go func() {
		for _, validator := range h.validators {
			proposalByt, err := proto.Marshal(propMsg)
			if err != nil {
				logrus.Error("smr::ProcessProposal decode proposal error: ", err)
				continue
			}
			_, err = h.env.P2pNetwork.RequestPeer(validator, ProposeCode, proposalByt)
			if err != nil {
				logrus.Errorf("smr::ProcessProposal request validator(%s) error: %v", validator.String(), err)
			} else {
				logrus.Debugf("smr:ProcessProposal::new proposal has been made, address(%s), proposal(%s)", h.LocalAddress(), utils.F(proposalID))
			}
		}
	}()
}
