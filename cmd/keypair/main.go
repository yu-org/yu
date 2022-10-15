package main

import (
	"fmt"
	"github.com/yu-org/yu/core/keypair"
	"os"
)

func main() {
	keyType := os.Args[1]
	secret := os.Args[2]
	pubkey, privkey, err := keypair.GenKeyPairWithSecret(keyType, []byte(secret))
	if err != nil {
		fmt.Println("generate keypair failed: ", err.Error())
		os.Exit(1)
	}
	fmt.Println("public key: ", pubkey.String())
	fmt.Println("public key with type: ", pubkey.StringWithType())
	fmt.Println("private key: ", privkey.String())
	fmt.Println("private key with type: ", privkey.StringWithType())
	fmt.Println("address: ", pubkey.Address().String())
}
