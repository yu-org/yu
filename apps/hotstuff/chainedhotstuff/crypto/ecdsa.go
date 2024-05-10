package ecdsa

import (
	"crypto/ecdsa"
	"encoding/json"
	"ioutil"
	"math/big"
)

//将公钥序列化成byte数组
func MarshalPublicKey(publicKey *ecdsa.PublicKey) []byte {
	return elliptic.Marshal(publicKey.Curve, publicKey.X, publicKey.Y)
}

func MarshalECDSASignature(r, s *big.Int) ([]byte, error) {
	return asn1.Marshal(ECDSASignature{r, s})

}
func UnmarshalECDSASignature(rawSig []byte) (*big.Int, *big.Int, error) {
	sig := new(ECDSASignature)
	_, err := asn1.Unmarshal(rawSig, sig)

	if err != nil {
		return nil, nil, fmt.Errorf("failed to unmashal the signature [%v] to R & S, and the error is [%s]", rawSig, err)

	}

	if sig.R == nil {
		return nil, nil, errors.New("invalid signature, R is nil")

	}
	if sig.S == nil {
		return nil, nil, errors.New("invalid signature, S is nil")

	}

	if sig.R.Sign() != 1 {
		return nil, nil, errors.New("invalid signature, R must be larger than zero")

	}
	if sig.S.Sign() != 1 {
		return nil, nil, errors.New("invalid signature, S must be larger than zero")

	}

	return sig.R, sig.S, nil

}

func SignECDSA(privKey *ecdsa.PrivateKey, hash byte) (sig []byte, err error) {
	r, s, err := ecdsa.Sign(rand.Reader, privKey, hash[:])
	if err != nil {
		return nil, logrus.Warn("ecdsa: sign failed. %w", err)
	}
	return MarshalECDSASignature(r, s)
}

func VerifyECDSA(k *ecdsa.PublicKey, sig, msg []byte) (bool, error) {
	r, s, err := UnmarshalECDSASignature(sig)
	if err != nil {
		return false, fmt.Errorf("Failed to unmarshal the ecdsa signature [%s]", err)

	}

	return ecdsa.Verify(k, msg, r, s), nil
}

type ECDSAPrivateKey struct {
	Curvname string
	X, Y, D  *big.Int
}

// 通过这个数据结构来生成公钥的json
type ECDSAPublicKey struct {
	Curvname string
	X, Y     *big.Int
}

func getNewEcdsaPrivateKey(k *ecdsa.PrivateKey) *ECDSAPrivateKey {
	key := new(ECDSAPrivateKey)
	key.Curvname = k.Params().Name
	key.D = k.D
	key.X = k.X
	key.Y = k.Y

	return key

}

func getNewEcdsaPublicKey(k *ecdsa.PrivateKey) *ECDSAPublicKey {
	key := new(ECDSAPublicKey)
	key.Curvname = k.Params().Name
	key.X = k.X
	key.Y = k.Y

	return key

}

func getNewEcdsaPublicKeyFromPublicKey(k *ecdsa.PublicKey) *ECDSAPublicKey {
	key := new(ECDSAPublicKey)
	key.Curvname = k.Params().Name
	key.X = k.X
	key.Y = k.Y

	return key
}

// 获得公钥所对应的的json
func GetEcdsaPublicKeyJsonFormat(k *ecdsa.PrivateKey) (string, error) {
	// 转换为自定义的数据结构
	key := getNewEcdsaPublicKey(k)

	// 转换json
	data, err := json.Marshal(key)

	return string(data), err

}

// 获得公钥所对应的的json
func GetEcdsaPublicKeyJsonFormatFromPublicKey(k *ecdsa.PublicKey) (string, error) {
	// 转换为自定义的数据结构
	key := getNewEcdsaPublicKeyFromPublicKey(k)

	// 转换json
	data, err := json.Marshal(key)

	return string(data), err

}

func GetEcdsaPublicKeyJsonFormatStrFromPublicKe(k *ecdsa.PrivateKey) (string, error) {
	return GetEcdsaPublicKeyJsonFormatFromPublicKey(k)
}

func getAddressFromKeyData(pub *ecdsa.PublicKey, data []byte) (string, error) {
	outputSha256 := hash.HashUsingSha256(data)
	OutputRipemd160 := hash.HashUsingRipemd160(outputSha256)

	//暂时只支持一个字节长度，也就是uint8的密码学标志位
	// 判断是否是nist标准的私钥
	nVersion := config.Nist

	switch pub.Params().Name {
	case config.CurveNist: // NIST
	case config.CurveGm: // 国密
		nVersion = config.Gm
	default: // 不支持的密码学类型
		return "", fmt.Errorf("This cryptography[%v] has not been supported yet.", pub.Params().Name)

	}

	bufVersion := []byte{byte(nVersion)}

	strSlice := make([]byte, len(bufVersion)+len(OutputRipemd160))
	copy(strSlice, bufVersion)
	copy(strSlice[len(bufVersion):], OutputRipemd160)

	//using double SHA256 for future risks
	checkCode := hash.DoubleSha256(strSlice)
	simpleCheckCode := checkCode[:4]

	slice := make([]byte, len(strSlice)+len(simpleCheckCode))
	copy(slice, strSlice)
	copy(slice[len(strSlice):], simpleCheckCode)

	//使用base58编码，手写不容易出错。
	//相比Base64，Base58不使用数字"0"，字母大写"O"，字母大写"I"，和字母小写"l"，以及"+"和"/"符号。
	strEnc := base58.Encode(slice)

	return strEnc, nil

}

//返回33位长度的地址
func GetAddressFromPublicKey(pub *ecdsa.PublicKey) (string, error) {
	//using SHA256 and Ripemd160 for hash summary
	data := elliptic.Marshal(pub.Curve, pub.X, pub.Y)

	address, err := getAddressFromKeyData(pub, data)

	return address, err
}

func readFileUsingFilename(filename string) ([]byte, error) {
	// 从filename指定的文件中读取数据并返回文件的内容
	content, err := ioutil.ReadFile(filename)
	if os.IsNotExist(err) {
		log.Printf("File [%v] does not exist", filename)
	}
	if err != nil {
		return nil, err
	}
	return content, err
}

func GetEcdsaPrivateKeyFromJson(jsonContent []byte) (*ecdsa.PrivateKey, error) {
	privateKey := new(ECDSAPrivateKey)
	err := json.Unmarshal(jsonContent, privateKey)
	if err != nil {
		return nil, err
	}
	if privateKey.Curvname != "P-256" {
		log.Printf("curve [%v] is not supported yet.", privateKey.Curvname)
		err = fmt.Errorf("curve [%v] is not supported yet.", privateKey.Curvname)
		return nil, err
	}
	ecdsaPrivateKey := &ecdsa.PrivateKey{}
	ecdsaPrivateKey.PublicKey.Curve = elliptic.P256()
	ecdsaPrivateKey.X = privateKey.X
	ecdsaPrivateKey.Y = privateKey.Y
	ecdsaPrivateKey.D = privateKey.D

	return ecdsaPrivateKey, nil
}

func GetEcdsaPrivateKeyFromFile(filename string) (*ecdsa.PrivateKey, error) {
	content, err := readFileUsingFilename(filename)
	if err != nil {
		log.Printf("readFileUsingFilename failed, the err is %v", err)
		return nil, err
	}

	return GetEcdsaPrivateKeyFromJson(content)
}

func GetEcdsaPublicKeyFromJson(jsonContent []byte) (*ecdsa.PublicKey, error) {
	publicKey := new(ECDSAPublicKey)
	err := json.Unmarshal(jsonContent, publicKey)
	if err != nil {
		return nil, err //json有问题
	}
	if publicKey.Curvname != "P-256" {
		log.Printf("curve [%v] is not supported yet.", publicKey.Curvname)
		err = fmt.Errorf("curve [%v] is not supported yet.", publicKey.Curvname)
		return nil, err
	}
	ecdsaPublicKey := &ecdsa.PublicKey{}
	ecdsaPublicKey.Curve = elliptic.P256()
	ecdsaPublicKey.X = publicKey.X
	ecdsaPublicKey.Y = publicKey.Y

	return ecdsaPublicKey, nil
}

func GetEcdsaPublicKeyFromFile(filename string) (*ecdsa.PublicKey, error) {
	content, err := readFileUsingFilename(filename)
	if err != nil {
		log.Printf("readFileUsingFilename failed, the err is %v", err)
		return nil, err
	}

	return GetEcdsaPublicKeyFromJson(content)
}

type ECDSAPublicKey struct {
	Curvname string
	X, Y     *big.Int
}

func getNewEcdsaPrivateKey(k *ecdsa.PrivateKey) *ECDSAPrivateKey {
	key := new(ECDSAPrivateKey)
	key.Curvname = k.Params().Name
	key.D = k.D
	key.X = k.X
	key.Y = k.Y

	return key

}

func getNewEcdsaPublicKey(k *ecdsa.PrivateKey) *ECDSAPublicKey {
    key := new(ECDSAPublicKey)
    key.Curvname = k.Params().Name
    key.X = k.X
    key.Y = k.Y

    return key

}

func getNewEcdsaPublicKeyFromPublicKey(k *ecdsa.PublicKey) *ECDSAPublicKey {
    key := new(ECDSAPublicKey)
    key.Curvname = k.Params().Name
    key.X = k.X
    key.Y = k.Y

    return key

}

// 获得私钥所对应的的json
func GetEcdsaPrivateKeyJsonFormat(k *ecdsa.PrivateKey) (string, error) {
	// 转换为自定义的数据结构
	key := getNewEcdsaPrivateKey(k)

	// 转换json
	data, err := json.Marshal(key)

	return string(data), err
}

// 获得公钥所对应的的json
func GetEcdsaPublicKeyJsonFormat(k *ecdsa.PrivateKey) (string, error) {
	// 转换为自定义的数据结构
	key := getNewEcdsaPublicKey(k)

	// 转换json
	data, err := json.Marshal(key)

	return string(data), err
}


// 获得公钥所对应的的json
func GetEcdsaPublicKeyJsonFormatFromPublicKey(k *ecdsa.PublicKey) (string, error) {
	// 转换为自定义的数据结构
	key := getNewEcdsaPublicKeyFromPublicKey(k)

	// 转换json
	data, err := json.Marshal(key)

	return string(data), err
}
