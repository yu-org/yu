package keypair

type KeyPair interface {
	SignData([]byte) ([]byte, error)

	VerifySigner(msg, sig []byte) (bool, error)
}

func GenKeyPair(keyType string) (KeyPair, error) {
	switch keyType {
	case "sr25519":
		return generateSr25519(), nil
	default:
		return generateEd25519()
	}
}
