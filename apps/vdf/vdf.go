package vdf

import (
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"math/big"
)

// ----------------------------
// 基础工具函数
// ----------------------------

// hashToPrime 从 N, x, y, t 哈希生成一个确定性的素数 l
func hashToPrime(N, x, y *big.Int, t uint64, bits int) (*big.Int, error) {
	h := sha256.New()
	h.Write(N.Bytes())
	h.Write(x.Bytes())
	h.Write(y.Bytes())

	tBytes := make([]byte, 8)
	for i := 0; i < 8; i++ {
		tBytes[7-i] = byte((t >> (8 * i)) & 0xff)
	}
	h.Write(tBytes)

	digest := h.Sum(nil)
	cand := new(big.Int).SetBytes(digest)

	if cand.BitLen() < bits {
		cand.Lsh(cand, uint(bits-cand.BitLen()))
	}
	if cand.Bit(0) == 0 {
		cand.Add(cand, big.NewInt(1))
	}

	for {
		if cand.ProbablyPrime(64) {
			return cand, nil
		}
		cand.Add(cand, big.NewInt(2))
	}
}

// gcd 计算 a,b 的最大公约数
func gcd(a, b *big.Int) *big.Int {
	return new(big.Int).GCD(nil, nil, a, b)
}

// GenerateRSAmodulus 生成一个 RSA 模数 N = p*q
// 注意：生产环境应使用多方安全生成，否则泄露因子会破坏安全性。
func GenerateRSAmodulus(totalBits int) (*big.Int, error) {
	if totalBits%2 != 0 {
		return nil, errors.New("bit length must be even")
	}
	half := totalBits / 2
	p, err := rand.Prime(rand.Reader, half)
	if err != nil {
		return nil, err
	}
	q, err := rand.Prime(rand.Reader, half)
	if err != nil {
		return nil, err
	}
	return new(big.Int).Mul(p, q), nil
}

// ----------------------------
// VDF Wesolowski 核心实现
// ----------------------------

// EvalResult 表示 VDF 的输出结果
type EvalResult struct {
	Y  *big.Int // y = x^{2^t} mod N
	Pi *big.Int // Wesolowski 证明 π
}

// Eval 计算 VDF 输出与证明
func Eval(x *big.Int, t uint64, N *big.Int, securityBits int) (*EvalResult, error) {
	xMod := new(big.Int).Mod(x, N)
	if xMod.Sign() == 0 {
		return nil, errors.New("x mod N == 0")
	}
	if gcd(xMod, N).Cmp(big.NewInt(1)) != 0 {
		return nil, errors.New("x not coprime with N")
	}

	// Step 1: 计算 y = x^{2^t} mod N
	y := new(big.Int).Set(xMod)
	for i := uint64(0); i < t; i++ {
		y.Mul(y, y)
		y.Mod(y, N)
	}

	// Step 2: 生成素数挑战 l
	l, err := hashToPrime(N, xMod, y, t, securityBits)
	if err != nil {
		return nil, err
	}

	// Step 3: q = floor(2^t / l), r = 2^t mod l
	twoPow := new(big.Int).Lsh(big.NewInt(1), uint(t))
	q := new(big.Int).Div(twoPow, l)
	// Step 4: π = x^q mod N
	pi := new(big.Int).Exp(xMod, q, N)

	return &EvalResult{Y: y, Pi: pi}, nil
}

// Verify 验证 Wesolowski 证明
func Verify(x, y, pi *big.Int, t uint64, N *big.Int, securityBits int) (bool, error) {
	xMod := new(big.Int).Mod(x, N)
	yMod := new(big.Int).Mod(y, N)
	piMod := new(big.Int).Mod(pi, N)

	l, err := hashToPrime(N, xMod, yMod, t, securityBits)
	if err != nil {
		return false, err
	}

	r := new(big.Int).Exp(big.NewInt(2), new(big.Int).SetUint64(t), l)
	left := new(big.Int).Exp(piMod, l, N)
	xr := new(big.Int).Exp(xMod, r, N)
	left.Mul(left, xr).Mod(left, N)

	return left.Cmp(yMod) == 0, nil
}
