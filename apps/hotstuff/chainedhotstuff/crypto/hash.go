import (
    "crypto/sha256"
    "crypto/sha512"
    "golang.org/x/crypto/ripemd160"
)

func DoubleSha256(data []byte) []byte {
    return HashUsingSha256(HashUsingSha256(data))
}

func HashUsingSha256(data []byte) []byte {
    h := sha256.New()
    h.Write(data)
    out := h.Sum(nil)

    return out
}

// Ripemd160，这种hash算法可以缩短长度
func HashUsingRipemd160(data []byte) []byte {
    h := ripemd160.New()
    h.Write(data)
    out := h.Sum(nil)

    return out

}
