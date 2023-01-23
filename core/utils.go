package core

import (
	"encoding/hex"
	"encoding/json"
	"time"

	c "../common"

	"github.com/golang/protobuf/proto"
	bpr "github.com/hyperledger/sawtooth-sdk-go/protobuf/batch_pb2"
	tpr "github.com/hyperledger/sawtooth-sdk-go/protobuf/transaction_pb2"
	sgn "github.com/hyperledger/sawtooth-sdk-go/signing"

	sha "crypto/sha512"
)

// address related constants used by various payloads
const (
	AddressLength   = 70
	NamespaceLength = 6
)

// VerifyPermission verify that initiator is permitted to transact in this transaction family
func VerifyPermission(familyName string, pubKey []byte) bool {
	// TODO implement this: should query state for permission
	// TODO logic to submit transaction to set state should be in banking/sysadmin
	// Note that we set permissions by family name NOT payload, why? It's perfectly logical to expect someone permissioned to set a rule, e.g., to also be permissioned to delete it

	return true // TODO// TODO// TODO// TODO// TODO// TODO// TODO
}

// VerifySignature signature is that of pl using pubKey
func VerifySignature(pl, signature, pubKey []byte) bool {
	sgnContext := sgn.CreateContext(EncryptionAlgoName)
	ok := sgnContext.Verify(signature, pl, sgn.NewSecp256k1PublicKey(pubKey))

	return ok
}

// CreateTransaction build a sawtooth transaction
func CreateTransaction(pl *c.SignedPayload, familyName string, inputs []string, outputs []string, dependencies []string) *tpr.Transaction {
	nonce := createNonce()
	bankPrivateKey, bankPubKey := getBankKeys()

	// json marshaling to get []byte which is needed in Transaction
	payloadBytes, err := json.Marshal(*pl)
	if err != nil {
		panic(err)
	}

	payloadSha512 := HexdigestB(payloadBytes)

	header := tpr.TransactionHeader{
		BatcherPublicKey: bankPubKey.AsHex(),
		Dependencies:     dependencies,
		FamilyName:       familyName,
		FamilyVersion:    FamilyVersion,
		Inputs:           inputs,
		Nonce:            nonce,
		Outputs:          outputs,
		PayloadSha512:    payloadSha512,
		SignerPublicKey:  bankPubKey.AsHex(),
	}

	headerBytes, err := proto.Marshal(&header)
	if err != nil {
		panic(err)
	}

	signer := c.GetSigner(bankPrivateKey)
	signedHeaderBytes := signer.Sign(headerBytes)
	signature := hex.EncodeToString(signedHeaderBytes)

	tx := tpr.Transaction{
		Header:          headerBytes,
		HeaderSignature: signature,
		Payload:         payloadBytes,
	}

	return &tx
}

// CheckLength the length required by sawtooth
func CheckLength(address string) string {
	// Note: arguably this is only needed in testing. once tests pass, no need to check in production. How to make it only used in testing?
	if len(address) != AddressLength {
		panic("wrong address length!")
	}

	return address
}

// HexdigestStr for strings
func HexdigestStr(str string) string {
	return HexdigestB([]byte(str))
}

// Namespace for state addresses
func Namespace(familyName string) string {
	return HexdigestStr(familyName)[:NamespaceLength]
}

// HashIt hash. Note: code taken from doSHA512 in signing
func HashIt(p []byte) []byte {
	hash := sha.New()
	hash.Write(p)

	return hash.Sum(nil)
}

// HexdigestB mimics hexdigest() in python's hashlib
func HexdigestB(p []byte) string {
	hashBytes := HashIt(p)
	return hex.EncodeToString(hashBytes)
}

func createNonce() string {
	ret := hex.EncodeToString([]byte(time.Now().String()))
	return ret
}

func getBankKeys() (bankPrivateKey sgn.PrivateKey, bankPubKey sgn.PublicKey) {
	bankPrivateKey, bankPubKey = c.GetKeysFromFiles(c.BatchSignerKeysFile)

	return

}

// GetBankAuthTools returns bank public key and signer object
func GetBankAuthTools() (bankPubKey sgn.PublicKey, signer *sgn.Signer) {
	privateKey, publicKey := getBankKeys()
	signer = c.GetSigner(privateKey)
	bankPubKey = publicKey

	return
}

func createBatchList(tx *tpr.Transaction) *bpr.BatchList {
	// bank signs all batches in our model
	signerKey, signer := GetBankAuthTools()

	txSignature := tx.HeaderSignature
	header := bpr.BatchHeader{
		SignerPublicKey: signerKey.AsHex(),
		TransactionIds:  []string{txSignature},
	}

	headerBytes, err := proto.Marshal(&header)
	if err != nil {
		panic(err)
	}

	signature := hex.EncodeToString(signer.Sign(headerBytes))
	batch := bpr.Batch{
		Header:          headerBytes,
		HeaderSignature: signature,
		Transactions:    []*tpr.Transaction{tx},
	}

	ret := &bpr.BatchList{Batches: []*bpr.Batch{&batch}}
	return ret
}
