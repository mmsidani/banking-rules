package core

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"

	"github.com/hyperledger/sawtooth-sdk-go/protobuf/transaction_pb2"

	c "../common"
	"github.com/hyperledger/sawtooth-sdk-go/processor"
	sgn "github.com/hyperledger/sawtooth-sdk-go/signing"
)

// PayloadClosePendingTx to cancel a tx pending while sigs are collected
type PayloadClosePendingTx c.PayloadClosePendingTx

// PayloadAddSigTx to add a sig to a pending tx requiring multi sigs
type PayloadAddSigTx c.PayloadAddSigTx

// PayloadListPendingTx for listing all pending tx requiring an initiator's sig
type PayloadListPendingTx c.PayloadListPendingTx

// PayloadSetPendingTx payload to set pending transaction and signature info. Note: this one is only defined here and not in ../common because a client never initiates a set pending tx
type PayloadSetPendingTx struct {
	SourceAccount   string
	BankTransaction []byte   // the bank transaction that requires multi sigs
	AuthorisedSigs  []string // List of pools (comma-separated lists) of authorised signers from the multisig rules that were triggered by this transaction
	RequiredMinSigs []int    // list of minimum required sigs: so, RequiredMinSigs[0] applies to the signers in AuthorisedSigs[0], a string with Transactor{} ID's separated by commas
	TransactionID   string
	Initiator       string // this is the initiator who initiated the query that resulted in this pending tx
}

// constants used in state address calculations
const (
	pendingNamespace    = "07"
	signatoriesSubspace = "08"
	transactionSubspace = "09"
	initiatorSubspace   = "12"
)

// PendingTxSigsInfo convenient structure to store required sigs info in the state
type PendingTxSigsInfo struct {
	AuthorisedSigs  []map[string]bool // one map per rule triggered. value true means signer signed
	RequiredMinSigs []int
}

// Apply applier for making a transaction pending
func (*PayloadSetPendingTx) Apply(pl []byte, context *processor.Context) error {
	var p PayloadSetPendingTx
	err := json.Unmarshal(pl, &p)
	if err != nil {
		panic(err)
	}

	root := pendingTxStateRootAddress(p.SourceAccount)
	txAddress := pendingTxTx(root, p.TransactionID)
	sigsAddress := pendingTxSigs(root, p.TransactionID)
	initiatorAddress := pendingTxInitiator(root, p.TransactionID)

	m := make(map[string][]byte)
	m[txAddress] = p.BankTransaction

	sigsInfo := initRequiredSigners(p.AuthorisedSigs, p.RequiredMinSigs, p.Initiator)
	sigsEnc, err := json.Marshal(*sigsInfo)
	if err != nil {
		panic(err)
	}
	m[sigsAddress] = sigsEnc

	m[initiatorAddress] = []byte(p.Initiator)

	addresses, err := context.SetState(m)
	if err != nil || len(addresses) != len(m) {
		return errors.New("error setting pending tx")
	}

	return nil
}

// Apply applier for closing a pending transaction. typical scenario is cancellation of the transaction
func (*PayloadClosePendingTx) Apply(pl []byte, context *processor.Context) error {
	var p PayloadClosePendingTx
	err := json.Unmarshal(pl, &p)
	if err != nil {
		panic(err)
	}

	// check if initator is the one who also initiated pending transaction that is to be cancelled
	pendingRootAddress := pendingTxStateRootAddress(p.SourceAccount)
	initiatorAddress := pendingTxInitiator(pendingRootAddress, p.TransactionID)
	m, err := context.GetState([]string{initiatorAddress})
	if err != nil {
		panic(err)
	}

	if string(m[initiatorAddress]) != p.Initiator {
		return &processor.InvalidTransactionError{Msg: "pending transaction can only be cancelled by party who initiated transaction"}
	}

	// check if pubkey is known to belong to initiator
	initiatorRootAddress := initiatorRootStateAddress(p.SourceAccount)
	pubKeysAddress := initiatorPubKeys(initiatorRootAddress, p.Initiator)
	m, err = context.GetState([]string{pubKeysAddress})
	initiatorPubKeys := make([]string, 0)
	err = json.Unmarshal(m[pubKeysAddress], &initiatorPubKeys)

	goodKey := checkKey(hex.EncodeToString(p.InitiatorKey), initiatorPubKeys)
	//////// TODO only keeping this for future tests of error prograpagation: this returns goodkey=false . goodKey := checkKey(string(p.InitiatorKey), initiatorPubKeys)
	if !goodKey {
		// TODO TODO TODO TODO decide what to do: panic or invalidtransactionerror?
		return &processor.InvalidTransactionError{Msg: "pubic key in transaction to cancel pending transaction is not recognised as a key for pending transaction initiator"}
	}

	// Note: instead of calculating the next 2 addresses, could I have just calculated the pending tx wild card and passed that to DeleteState()?
	txAddress := pendingTxTx(pendingRootAddress, p.TransactionID)
	sigsAddress := pendingTxSigs(pendingRootAddress, p.TransactionID)
	addresses := []string{txAddress, sigsAddress, initiatorAddress}

	delAddresses, err := context.DeleteState(addresses)
	if err != nil || len(delAddresses) != len(addresses) {
		// TODO change return type to error code
		///////////////////////////////////////////////////return errors.New("error deleting pending tx")
		panic("deladdresses")
	}

	return nil
}

// Apply applier for adding signatures to a pending transaction
func (*PayloadAddSigTx) Apply(pl []byte, context *processor.Context) error {
	var p PayloadAddSigTx
	err := json.Unmarshal(pl, &p)
	if err != nil {
		panic(err)
	}

	pendingTxRootAddress := pendingTxStateRootAddress(p.SourceAccount)
	txAddress := pendingTxTx(pendingTxRootAddress, p.TransactionID)
	sigsAddress := pendingTxSigs(pendingTxRootAddress, p.TransactionID)

	initiatorRootAddress := initiatorRootStateAddress(p.SourceAccount)
	pubKeysAddress := initiatorPubKeys(initiatorRootAddress, p.Initiator)

	m, err := context.GetState([]string{txAddress, sigsAddress, pubKeysAddress})
	if err != nil {
		panic(err)
	}

	// TODO check tx was not closed before this initiator got to sign it
	var sigsInfo PendingTxSigsInfo
	err = json.Unmarshal(m[sigsAddress], &sigsInfo)
	if err != nil {
		panic(err)
	}

	// check initiator
	indices := checkInitiator(p.Initiator, sigsInfo.AuthorisedSigs)
	if indices == nil {
		return &processor.InvalidTransactionError{Msg: p.Initiator + " not authorised to sign or has already signed transaction" + p.TransactionID}
	}

	// check initiator key. Note: to avoid hard coding a structure for the public keys we avoid unmarshaling here. otherwise, code here will break if structure (which as of today 8/17/18 is []string) is changed in the pubkeys applier. When does this break?
	strKey := hex.EncodeToString(p.PubKey)
	foundKey := strings.Contains(string(m[pubKeysAddress]), strKey)
	if !foundKey {
		return &processor.InvalidTransactionError{Msg: strKey + " is not recognised as a public key for " + p.Initiator}
	}

	// now we verify the signature. we expect the signature in the payload to be that of the stored, pending tx (a marshaled PayloadQueryAuth as of 9/6/18)
	sgnContext := sgn.CreateContext(EncryptionAlgoName)
	ok := sgnContext.Verify(p.Signature, m[txAddress], sgn.NewSecp256k1PublicKey(p.PubKey))
	if !ok {
		panic("Invalid signature for pending transaction " + p.TransactionID)
	}

	// OK so we have a legit Initiator with a legit key with a legit sig. so we map signer to true in the sigsInfo structure and decrement the number of required signatures
	removeInitiator(p.Initiator, &sigsInfo, indices)
	sigsEnc, err := json.Marshal(sigsInfo)
	if err != nil {
		panic(err)
	}

	addresses, err := context.SetState(map[string][]byte{sigsAddress: sigsEnc})
	if err != nil || len(addresses) == 0 {
		return &processor.InvalidTransactionError{Msg: "error updating signatures for transaction " + p.TransactionID}
	}
	// check if more sigs are required. Note that we update the state before possibly deleting it so we have a record of all signatures
	moreSigs := checkRemainingSigs(sigsInfo.RequiredMinSigs)
	if !moreSigs {
		// leaves we haven't concerned ourselves with yet
		initiatorAddress := pendingTxInitiator(pendingTxRootAddress, p.TransactionID)

		addresses := []string{txAddress, sigsAddress, initiatorAddress}
		delAddresses, err := context.DeleteState(addresses)
		if err != nil || len(delAddresses) != len(addresses) {
			panic(err)
		}
	}

	return nil
}

// Handle return all payloads of all pending transactions awaiting payload.Initiator's signature
func (*PayloadListPendingTx) Handle(pl []byte) map[string]interface{} {
	var p PayloadListPendingTx
	err := json.Unmarshal(pl, &p)
	if err != nil {
		panic(err)
	}

	pendingTxRootAddress := pendingTxStateRootAddress(p.SourceAccount)
	rootSigsAddr := pendingTxSigsWildCard(pendingTxRootAddress)
	sigsAddresses, sigsAllTx := SubmitStateReq(rootSigsAddr)
	txIds := make([]string, 0)
	for i, signersPerTx := range sigsAllTx {
		var sigsInfo PendingTxSigsInfo
		err = json.Unmarshal(signersPerTx, &sigsInfo)
		if err != nil {
			panic(err)
		}

		for _, signersPerRule := range sigsInfo.AuthorisedSigs {
			if v, _ := signersPerRule[p.Initiator]; !v {
				// get the ending string that was used in building the address of the pending tx
				txIds = append(txIds, strings.Replace(sigsAddresses[i], rootSigsAddr, "", 1))
			}
		}
	}
	ret := make(map[string]interface{}, 0)
	if len(txIds) == 0 {
		return ret
	}

	rootPendingTxAddr := pendingTxTxWildCard(pendingTxRootAddress)
	for _, uid := range txIds {
		address := rootPendingTxAddr + uid
		_, ptx := SubmitStateReq(address)

		var px PayloadQueryAuth
		err = json.Unmarshal(ptx[0], &px)
		if err != nil {
			panic(err)
		}

		ret[uid] = px
	}

	return ret
}

// Handle add sig tx. Note: add sig tx is a state changing request. Unlike other state changing requests however which just have to make sure the transaction was committed (through SubmitTx), add sig tx needs to know if all sigs have been obtained. Since the Apply() method invoked from the validator has to return error only, I added a Handle() method which checks to see if more sigs are still needed after this sig has been added
func (*PayloadAddSigTx) Handle(pl []byte) map[string]interface{} {
	var p PayloadAddSigTx
	err := json.Unmarshal(pl, &p)
	if err != nil {
		panic(err)
	}

	// create signedPayload (recreate really since this is called from a point, lambda handler, where pl is wrapped in SignedPayload. not worried about cost. and no easy way to solve the inelegant .inefficiency)
	bankPubKey, signer := GetBankAuthTools()
	signature := signer.Sign(pl)
	signedPayload := c.SignedPayload{
		Type:         "add_sig_tx",
		SignerPubKey: bankPubKey.AsBytes(),
		Signature:    signature,
		Payload:      pl,
	}

	tx := (&PayloadAddSigTx{}).WrapInTx(&signedPayload)
	_ = SubmitTx(tx)

	// SubmitTx polls until transaction has been committed. So, if we're here, the sig has been added and the state updated and we need to know if all sigs are in. we do that by checking if the address where the pending tx was stored is still valid because the Apply() method on *PayloadAddSigTx deletes the state after all sigs are in.

	pendingAddress := pendingTxTx(pendingTxStateRootAddress(p.SourceAccount), p.TransactionID)
	a, _ := SubmitStateReq(pendingAddress)
	if a == nil || len(a) == 0 {
		// state leaf for the pending transaction was deleted
		return map[string]interface{}{"action": "allow"}
	}

	// we're here therefore more sigs are still needed
	return map[string]interface{}{"action": "pending"}
}

// Note PayloadSetPendingTx does NOT need a WrapInTx() method because it's never initiated by the client

// WrapInTx signedPayload with payloadClosePendingTx to submit to validator
func (*PayloadClosePendingTx) WrapInTx(pl *c.SignedPayload) *transaction_pb2.Transaction {
	ok := VerifySignature(pl.Payload, pl.Signature, pl.SignerPubKey)
	if !ok {
		panic("Invalid signature for close pending transaction ")
	}

	var p PayloadClosePendingTx
	err := json.Unmarshal(pl.Payload, &p)
	if err != nil {
		panic(err)
	}

	pendingRootAddress := pendingTxStateRootAddress(p.SourceAccount)
	initiatorAddress := pendingTxInitiator(pendingRootAddress, p.TransactionID)
	txAddress := pendingTxTx(pendingRootAddress, p.TransactionID)
	sigsAddress := pendingTxSigs(pendingRootAddress, p.TransactionID)

	initiatorRootAddress := initiatorRootStateAddress(p.SourceAccount)
	pubKeysAddress := initiatorPubKeys(initiatorRootAddress, p.Initiator)

	inputs := []string{txAddress, sigsAddress, initiatorAddress, pubKeysAddress}
	outputs := []string{txAddress, sigsAddress, initiatorAddress}
	dependencies := []string{}

	fn := p.SourceAccount
	ok = VerifyPermission(fn, pl.SignerPubKey)
	if !ok {
		panic("signer of close pending transaction is not authorised")
	}

	return CreateTransaction(pl, fn, inputs, outputs, dependencies)
}

// WrapInTx signedPayload with PayloadAddSigTx to submit to validator
func (*PayloadAddSigTx) WrapInTx(pl *c.SignedPayload) *transaction_pb2.Transaction {
	ok := VerifySignature(pl.Payload, pl.Signature, pl.SignerPubKey)
	if !ok {
		panic("Invalid signature for add sig transaction ")
	}

	var p PayloadAddSigTx
	err := json.Unmarshal(pl.Payload, &p)
	if err != nil {
		panic(err)
	}

	pendingRootAddress := pendingTxStateRootAddress(p.SourceAccount)
	initiatorAddress := pendingTxInitiator(pendingRootAddress, p.TransactionID)
	txAddress := pendingTxTx(pendingRootAddress, p.TransactionID)
	sigsAddress := pendingTxSigs(pendingRootAddress, p.TransactionID)

	initiatorRootAddress := initiatorRootStateAddress(p.SourceAccount)
	pubKeysAddress := initiatorPubKeys(initiatorRootAddress, p.Initiator)

	inputs := []string{txAddress, sigsAddress, initiatorAddress, pubKeysAddress}
	outputs := []string{txAddress, sigsAddress, initiatorAddress}
	dependencies := []string{}

	fn := p.SourceAccount
	ok = VerifyPermission(fn, pl.SignerPubKey)
	if !ok {
		panic("signer of add sig transaction is not authorised")
	}

	return CreateTransaction(pl, fn, inputs, outputs, dependencies)
}

func checkInitiator(initiator string, authorisedSigs []map[string]bool) []int {
	ret := make([]int, 0)
	for i, item := range authorisedSigs {
		if v, _ := item[initiator]; !v {
			ret = append(ret, i) // we're here if v = false, i.e., initiator has NOT signed
		}
	}

	if len(ret) == 0 {
		return nil
	}

	return ret
}

// no return because sigsInfo is modified in here
func removeInitiator(initiator string, sigsInfo *PendingTxSigsInfo, indices []int) {
	for _, index := range indices {
		sigsInfo.AuthorisedSigs[index][initiator] = true // this means initiator has already signed
		sigsInfo.RequiredMinSigs[index]--
	}
}

// check if all entries in numSigsLeft are 0. Implicitly assumes that entries in numSigsLeft are non-negative. TODO must assert that somewhere else
func checkRemainingSigs(numSigsLeft []int) bool {
	for _, item := range numSigsLeft {
		if item != 0 {
			return true // more signatures are needed for this pending transactions
		}
	}
	return false // requirements from all rules have been met
}

// check if key matches an entry in pubKeys
func checkKey(key string, pubKeys []string) bool {
	for _, k := range pubKeys {
		if k == key {
			return true
		}
	}

	return false
}

// signers is an array of comma-separated list of required signers
func initRequiredSigners(signers []string, minSigs []int, initiator string) *PendingTxSigsInfo {
	var ret PendingTxSigsInfo
	ret.AuthorisedSigs = make([]map[string]bool, len(signers))
	ret.RequiredMinSigs = minSigs
	for i, ruleSigners := range signers {
		ret.AuthorisedSigs[i] = make(map[string]bool)
		s := strings.Split(ruleSigners, ",")
		for _, signer := range s {
			if signer != initiator {
				ret.AuthorisedSigs[i][signer] = false // not signed yet
			} else {
				ret.AuthorisedSigs[i][initiator] = true // initiator submitted transaction so obviously approves it
				ret.RequiredMinSigs[i]--
			}
		}
	}

	return &ret
}

// // // build a set of out the list of lists of required signers. every entry in signersLists is a string containing a comma separated list of signers (no set structure in golang so we use map)
// // func getRequiredSigners(signersLists []string) map[string]struct{} {
// // 	m := make(map[string]struct{})
// // 	for _, signers := range signersLists {
// // 		signersArray := strings.Split(signers, ",")
// // 		for _, signer := range signersArray {
// // 			m[signer] = struct{}{}
// // 		}
// // 	}

// // 	return m
// // }

type pendingRootAddressType string

func pendingTxStateRootAddress(sourceAccount string) pendingRootAddressType {
	// NO permission tag for pending transactions logic
	return pendingRootAddressType(Namespace(familyName(sourceAccount, "")) + pendingNamespace)
}

// wild card for all pending Tx and their subspaces
func pendingTxWildCard(root pendingRootAddressType) string {
	return string(root)
}

// root address of all required signatories for all pending tx
func pendingTxSigsWildCard(root pendingRootAddressType) string {
	return pendingTxWildCard(root) + signatoriesSubspace
}

// address of required sigs and those already satisfied for uid. Note: uid is any unique identifier
func pendingTxSigs(root pendingRootAddressType, uid string) string {
	return CheckLength(pendingTxSigsWildCard(root) + formatPendingTxUID(uid))
}

// // // // // the root of all addresses of all required signers for all pending transactions. The intent is to help looking up efficiently whether an initiator's signature is needed for any pending transaction
// // // // func pendingRequiredSignersWildCard(root pendingRootAddressType) string {
// // // // 	return pendingTxWildCard(root) + requiredSignersSubspace
// // // // }

// // // // // the address where we store the union of all authorised/required signers for uid
// // // // func pendingRequiredSigners(root pendingRootAddressType, uid string) string {
// // // // 	return CheckLength(pendingRequiredSignersWildCard(root) + formatPendingTxUID(uid))
// // // // }

// root address of all addresses of pending transactions structs
func pendingTxTxWildCard(root pendingRootAddressType) string {
	return pendingTxWildCard(root) + transactionSubspace
}

// address of the actual (the text) of the transaction. see also previous comment
func pendingTxTx(root pendingRootAddressType, uid string) string {
	return CheckLength(pendingTxTxWildCard(root) + formatPendingTxUID(uid))
}

// root address of all initiators of pending transactions
func pendingTxInitiatorWildCard(root pendingRootAddressType) string {
	return pendingTxWildCard(root) + initiatorSubspace
}

// address of the initiator of the pending tx
func pendingTxInitiator(root pendingRootAddressType, uid string) string {
	return CheckLength(pendingTxInitiatorWildCard(root) + formatPendingTxUID(uid))
}

func formatPendingTxUID(uid string) string {
	return uid[:actorLength+fieldLength]
}
