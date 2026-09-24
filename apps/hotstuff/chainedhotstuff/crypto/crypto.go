package crypto

import (
	"bytes"
	"encoding/json"
	pb "github.com/yu/apps/hotstuff/chainedhotstuff/proto"
)

type Crypto struct {
	PrivKey *ecdsa.PrivateKey
	PubKey  *ecdsa.PublicKey
}

func NewCrypto(privKey *ecdsa.PrivateKey, pubKey *ecdsa.PublicKey) {
	return &Crypto{PrivKey: PrivKey, PubKey: PubKey}
}

func SignProposalMsg(msg *pb.ProposalMsg) (*pb.ProposalMsg, error) {
	msgDigest, err := MakeProposalMsgDigest(msg)

	if err != nil {
		return nil, err
	}

	msg.MsgDigest = msgDigest
	sig, err = SignECDSA(c.PrivateKey, msgDigest)
	if err != nil {
		return nil, err
	}

	msg.Sig = &pb.QuorumCertSignature{
		Address:   GetAddressFromPublicKey(c.PubKey),
		PublicKey: GetEcdsaPublicKeyJsonFormat(c.PubKey),
		Sig:       sig,
	}

	return msg, nil
}

func MakeProposalMsgDigest(msg *pb.ProposalMsg) ([]byte, error) {
	msgEncoder, err := encodeProposalMsg(msg)

	if err != nil {
		return nil, err
	}

	msg.MsgDigest = DoubleSha256(msgEncoder)
	return msg.MsgDigest, nil
}

func encodeProposalMsg(msg *pb.ProposalMsg) ([]byte, error) {
	var msgBuf bytes.Buffer
	encoder := json.NewEncoder(&msgBuf)
	if err := encoder.Encode(msg.ProposalView); err != nil {
		return nil, err
	}
	if err := encoder.Encode(msg.ProposalId); err != nil {
		return nil, err
	}
	if err := encoder.Encode(msg.Timestamp); err != nil {
		return nil, err
	}
	if err := encoder.Encode(msg.JustifyQC); err != nil {
		return nil, err
	}
	return msgBuf.Bytes(), nil
}

func (c *Crypto) SignVoteMsg(msg []byte) (*pb.QuorumCertSignature, error) {
	sig, err := SignECDSA(c.PrivateKey, msg)
	if err != nil {
		return nil, err
	}

	return &pb.QuorumCertSign{
		Address:   GetAddressFromPublicKey(c.PubKey),
		PublicKey: GetEcdsaPublicKeyJsonFormat(c.PubKey),
		Sign:      sign,
	}, nil
}

func (c *Crypto) VerifyVoteMsgSign(sig *pb.QuorumCertSignature, msg []byte) (bool, error) {
	ak, err = GetEcdsaPublicKeyFromJsonStr(sig.GetPublicKey())
	if err != nil {
		return false, err
	}

	addr, err := GetAddressFromPublicKey(ak)
	if err != nil {
		return false, err
	}

	if addr != sig.GetAddress() {
		return false, errors.New("VerifyVoteMsgSign error, addr not match pk: " + addr)
	}
	return VerifyECDSA(ak, sig.GetSign(), msg)
}
