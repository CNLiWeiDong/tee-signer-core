package secp256r1

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/crypto"
)

func GenerateKeypair() (string, string, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", "", err
	}
	bprk := crypto.FromECDSA(key)
	pub := key.PublicKey
	bpuk := elliptic.Marshal(elliptic.P256(), pub.X, pub.Y) // compressed key  crypto.FromECDSAPub(&pub)
	return hex.EncodeToString(bprk), hex.EncodeToString(bpuk), nil
}

func HexToPub(hexPub string) (*ecdsa.PublicKey, error) {
	bpuk, err := hex.DecodeString(hexPub)
	if err != nil {
		return nil, err
	}
	x, y := elliptic.Unmarshal(elliptic.P256(), bpuk)
	if x == nil {
		return nil, errors.New("invalid secp256k1 public key")
	}
	return &ecdsa.PublicKey{Curve: elliptic.P256(), X: x, Y: y}, nil
}

func HexToPri(hexPri string) (*ecdsa.PrivateKey, error) {
	bprk, err := hex.DecodeString(hexPri)
	if err != nil {
		return nil, err
	}
	priv := new(ecdsa.PrivateKey)
	priv.PublicKey.Curve = elliptic.P256()
	priv.D = new(big.Int).SetBytes(bprk)

	priv.PublicKey.X, priv.PublicKey.Y = priv.PublicKey.Curve.ScalarBaseMult(bprk)
	if priv.PublicKey.X == nil {
		return nil, errors.New("invalid private key")
	}
	return priv, nil
}

func Sign(data []byte, hexPri string) ([]byte, error) {
	prk, err := HexToPri(hexPri)
	if err != nil {
		return nil, err
	}
	r, s, err := ecdsa.Sign(rand.Reader, prk, data)
	if err != nil {
		return nil, err
	}
	params := prk.Curve.Params()
	curveOrderByteSize := params.P.BitLen() / 8
	rBytes, sBytes := r.Bytes(), s.Bytes()
	signature := make([]byte, curveOrderByteSize*2)
	copy(signature[curveOrderByteSize-len(rBytes):], rBytes)
	copy(signature[curveOrderByteSize*2-len(sBytes):], sBytes)
	return signature, nil
}

func Verify(data, signature []byte, hexPub string) bool {
	puk, err := HexToPub(hexPub)
	if err != nil {
		return false
	}
	curveOrderByteSize := puk.Curve.Params().P.BitLen() / 8
	r, s := new(big.Int), new(big.Int)
	r.SetBytes(signature[:curveOrderByteSize])
	s.SetBytes(signature[curveOrderByteSize:])
	return ecdsa.Verify(puk, data, r, s)
}
