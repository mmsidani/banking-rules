package common

import (
	"bufio"
	"encoding/hex"
	"errors"
	"os"
	"reflect"

	sgn "github.com/hyperledger/sawtooth-sdk-go/signing"
)

const (
	// TODO these 2 belong in config.go?
	pubKeyFileSuffix = "pub"
	// private key file suffix
	privKeyFileSuffix = "priv"
)

// Struct2Map convert a struct to a map. r MUST be a pointer to a struct.
func Struct2Map(r interface{}) map[string]interface{} {
	s := reflect.ValueOf(r).Elem()
	m := make(map[string]interface{}, 0)
	typeOfR := s.Type()
	for i := 0; i < s.NumField(); i++ {
		m[typeOfR.Field(i).Name] = s.Field(i).Interface()
	}

	return m
}

// GetKeysFromFiles reads keys from file
func GetKeysFromFiles(keysFilePrefix string) (sgn.PrivateKey, sgn.PublicKey) {
	pubKeyFileName := keysFilePrefix + "." + pubKeyFileSuffix
	privKeyFileName := keysFilePrefix + "." + privKeyFileSuffix

	privateKey := readOneLineHex(privKeyFileName)
	publicKey := readOneLineHex(pubKeyFileName)

	return sgn.NewSecp256k1PrivateKey(privateKey), sgn.NewSecp256k1PublicKey(publicKey)
}

// GetPubKeyFromFile reads one key file
func GetPubKeyFromFile(fileName string) sgn.PublicKey {
	return sgn.NewSecp256k1PublicKey(readOneLineHex(fileName))
}

// GetSigner from signing package: a struct wrapping a context and a private key
func GetSigner(privateKey sgn.PrivateKey) *sgn.Signer {
	return sgn.NewCryptoFactory(sgn.CreateContext(privateKey.GetAlgorithmName())).NewSigner(privateKey)
}

func readOneLineHex(fileName string) []byte {
	f, err := os.Open(fileName)
	defer f.Close()
	if err != nil {
		panic(errors.New("Error opening keys file"))
	}

	// file has one line only: do I still need Split()? (not crucial)
	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)

	ok := scanner.Scan()
	if !ok {
		panic("error reading file")
	}

	ret, err := hex.DecodeString(scanner.Text())
	if err != nil {
		panic(err)
	}

	return ret
}
