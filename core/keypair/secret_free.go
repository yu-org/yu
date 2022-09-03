package keypair

import . "github.com/yu-org/yu/common"

type FreePubkey struct{}

func (s *FreePubkey) Type() string {
	return SecretFree
}

func (s *FreePubkey) Equals(key Key) bool {
	_, ok := key.(*FreePubkey)
	return ok
}

func (s *FreePubkey) Bytes() []byte {
	return nil
}

func (s *FreePubkey) String() string {
	return ToHex(s.Bytes())
}

func (s *FreePubkey) BytesWithType() []byte {
	return append([]byte(SecretFreeIdx), s.Bytes()...)
}

func (s *FreePubkey) StringWithType() string {
	return ToHex(s.BytesWithType())
}

func (s *FreePubkey) Address() Address {
	return NullAddress
}

func (s *FreePubkey) VerifySignature(msg, sig []byte) bool {
	return true
}

// ------ Private key ------

type FreePrivkey struct{}

func (f *FreePrivkey) Type() string {
	return SecretFree
}

func (f *FreePrivkey) Equals(key Key) bool {
	_, ok := key.(*FreePrivkey)
	return ok
}

func (f *FreePrivkey) Bytes() []byte {
	return nil
}

func (f *FreePrivkey) String() string {
	return ToHex(f.Bytes())
}

func (f *FreePrivkey) BytesWithType() []byte {
	return append([]byte(SecretFreeIdx), f.Bytes()...)
}

func (f *FreePrivkey) StringWithType() string {
	return ToHex(f.BytesWithType())
}

func (f *FreePrivkey) SignData(bytes []byte) ([]byte, error) {
	return nil, nil
}
