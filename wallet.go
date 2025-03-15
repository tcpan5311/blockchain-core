package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"log"
	"math/big"

	"golang.org/x/crypto/ripemd160"
)

const version = byte(0x00)
const walletFile = "wallet.dat"
const addressChecksumLen = 4

// Wallet stores private and public keys
type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

// NewWallet creates and returns a Wallet
func NewWallet() *Wallet {
	private, public := newKeyPair()
	wallet := Wallet{*private, public}

	return &wallet
}

// GetAddress returns wallet address
func (w Wallet) GetAddress() []byte {
	pubKeyHash := HashPubKey(w.PublicKey)

	versionedPayload := append([]byte{version}, pubKeyHash...)
	checksum := checksum(versionedPayload)

	fullPayload := append(versionedPayload, checksum...)
	address := Base58Encode(fullPayload)

	return address
}

// HashPubKey hashes public key
func HashPubKey(pubKey []byte) []byte {
	publicSHA256 := sha256.Sum256(pubKey)

	RIPEMD160Hasher := ripemd160.New()
	_, err := RIPEMD160Hasher.Write(publicSHA256[:])
	if err != nil {
		log.Panic(err)
	}
	publicRIPEMD160 := RIPEMD160Hasher.Sum(nil)

	return publicRIPEMD160
}

// ValidateAddress checks if an address is valid
func ValidateAddress(address string) bool {
	pubKeyHash := Base58Decode([]byte(address))
	actualChecksum := pubKeyHash[len(pubKeyHash)-addressChecksumLen:]
	version := pubKeyHash[0]
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-addressChecksumLen]
	targetChecksum := checksum(append([]byte{version}, pubKeyHash...))

	return bytes.Compare(actualChecksum, targetChecksum) == 0
}

// checksum generates a checksum for a public key
func checksum(payload []byte) []byte {
	firstSHA := sha256.Sum256(payload)
	secondSHA := sha256.Sum256(firstSHA[:])

	return secondSHA[:addressChecksumLen]
}

func newKeyPair() (*ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()
	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panic(err)
	}
	pubKey := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)

	return private, pubKey
}

// Custom serialization methods to handle ecdsa.PrivateKey
type walletGob struct {
	D, X, Y []byte
	PublicKey []byte
}

func (w *Wallet) GobEncode() ([]byte, error) {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	data := walletGob{
		D:        w.PrivateKey.D.Bytes(),
		X:        w.PrivateKey.PublicKey.X.Bytes(),
		Y:        w.PrivateKey.PublicKey.Y.Bytes(),
		PublicKey: w.PublicKey,
	}

	err := encoder.Encode(data)
	if err != nil {
		return nil, err
	}

	return result.Bytes(), nil
}

func (w *Wallet) GobDecode(data []byte) error {
	decoder := gob.NewDecoder(bytes.NewReader(data))
	var decoded walletGob

	err := decoder.Decode(&decoded)
	if err != nil {
		return err
	}

	curve := elliptic.P256()
	w.PrivateKey = ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{
			Curve: curve,
			X:     new(big.Int).SetBytes(decoded.X),
			Y:     new(big.Int).SetBytes(decoded.Y),
		},
		D: new(big.Int).SetBytes(decoded.D),
	}

	w.PublicKey = decoded.PublicKey
	return nil
}
