package ed25519

import (
	"crypto"
	"crypto/ed25519"
	crypto_rand "crypto/rand"
	"encoding/hex"
)

func GenerateKeypair() (string, string, error) {
	pub, prv, err := ed25519.GenerateKey(crypto_rand.Reader)
	if err != nil {
		return "", "", err
	}
	return hex.EncodeToString(prv), hex.EncodeToString(pub), nil
}

func HexToPri(priHex string) (ed25519.PrivateKey, error) {
	seedBytes, err := hex.DecodeString(priHex)
	if err != nil {
		return nil, err
	}
	return ed25519.NewKeyFromSeed(seedBytes), nil
}

func HexToPub(pubHex string) (ed25519.PublicKey, error) {
	pubBytes, err := hex.DecodeString(pubHex)
	if err != nil {
		return nil, err
	}
	return ed25519.PublicKey(pubBytes), nil
}

func Sign(priHex string, message []byte) (signature []byte, err error) {
	privateKey, err := HexToPri(priHex)
	if err != nil {
		return nil, err
	}
	return privateKey.Sign(crypto_rand.Reader, message, crypto.Hash(0))
}

func Verify(pubHex string, message, sig []byte) bool {
	pub, err := HexToPub(pubHex)
	if err != nil {
		return false
	}
	return ed25519.Verify(pub, message, sig)
}
