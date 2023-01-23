package common

import (
	"encoding/hex"
	"encoding/json"
	"strings"

	sgn "github.com/hyperledger/sawtooth-sdk-go/signing"
)

// PayloadFields is for parsing command line arguments
type PayloadFields struct {
	// TODO TODO TODO TODO KeysFile doesn't belong here
	KeysFile      string  `short:"k" long:"keyfile" description:"Keys file"`
	RequestType   string  `short:"t" long:"type" description:"request type"`
	Rule          string  `short:"r" long:"rule" description:"the rule being set"`
	RuleHash      string  `short:"h" long:"rulehash" description:"the hash of the rule being set or deleted"`
	Initiator     string  `short:"i" long:"actioninitiator" description:"initiator of transaction on account"`
	Action        string  `short:"a" long:"action" description:"action on account, withdrawal, transfer, ..."`
	Recipient     string  `short:"e" long:"recipient" description:"the beneficiary of the transaction on account"`
	Amount        float64 `short:"m" long:"amount" description:"amount to be withdrawn, transferred"`
	SourceAccount string  `short:"s" long:"sourceaccount" description:"the account from which amount is withdrawn,..."`
	DestAccount   string  `short:"d" long:"destaccount" description:"beneficiary of payment, transfer, ..."`
	Group         string  `short:"g" long:"group" description:"label for a group of initiators. name must end with "`
	PubKeys       string  `long:"pubkeys" description:"(comma-separated) public keys to be associated with initiator"`
	Signature     string  `long:"signature" description:"signature for a pending transaction"`
	TransactionID string  `long:"transactionid" description:"system generated id displayed to user"`
	InitiatorKey  string  `long:"initiatorkey" description:"the initiator public key"`
}

// Important Note: this should have every type of payload
var payloadToCreatorMethod = map[string]func(*map[string]interface{}) []byte{
	"set_initiator_rule":          setInitiatorRule,
	"delete_initiator_rule":       deleteInitiatorRule,
	"list_initiator_rules":        listInitiatorRules,
	"add_initiator_to_group":      addInitiatorToGroup,
	"remove_initiator_from_group": removeInitiatorFromGroup,
	"list_initiator_groups":       listInitiatorGroups,
	"set_initiator_pub_keys":      setInitiatorPubKeys,
	"delete_initiator_pub_keys":   deleteInitiatorPubKeys,
	"list_initiator_pub_keys":     listInitiatorPubKeys,
	"query_auth":                  queryAuth,
	"close_pending_tx":            closePendingTx,
	"add_sig_tx":                  addSigTx,
	"list_pending_tx":             listPendingTx,
	"set_recipient":               setRecipient,
	"remove_recipient":            removeRecipient,
	"list_recipient":              listRecipient,
	"set_account_level_rule":      setAccountLevelRule,
	"delete_account_level_rule":   deleteAccountLevelRule,
	"list_account_level_rules":    listAccountLevelRules,
}

// CreateSignedPayload (typically) from command line arguments
func CreateSignedPayload(opts *PayloadFields, signerPrivateKey *sgn.PrivateKey, signerPubKey *sgn.PublicKey) *SignedPayload {
	m := Struct2Map(opts)

	// Important Note: closing pending tx requires initiator key which is not set on the command line. so we add it here. Note that signerPubKey, normally, does not have to be a key of the initiator from the command line. But for closing pending transactions we require that.
	m["InitiatorKey"] = (*signerPubKey).AsBytes()

	payloadType := opts.RequestType

	creator, ok := payloadToCreatorMethod[payloadType]
	if !ok {
		panic("unknown request type")
	}

	p := creator(&m)

	signer := GetSigner(*signerPrivateKey)
	signature := signer.Sign(p)

	return &SignedPayload{SourceAccount: opts.SourceAccount, Type: payloadType, SignerPubKey: (*signerPubKey).AsBytes(), Signature: signature, Payload: p}
}

func setInitiatorRule(mp *map[string]interface{}) []byte {
	m := *mp

	a := m["SourceAccount"].(string)
	i := m["Initiator"].(string)
	r := m["Rule"].(string)

	payload := PayloadSetInitiatorRule{
		SourceAccount: a,
		Initiator:     i,
		Rule:          r,
	}

	pEnc, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}

	return pEnc
}

func deleteInitiatorRule(mp *map[string]interface{}) []byte {
	m := *mp

	a := m["SourceAccount"].(string)
	i := m["Initiator"].(string)
	r := m["RuleHash"].(string)

	payload := PayloadDeleteInitiatorRule{
		SourceAccount: a,
		Initiator:     i,
		RuleHash:      r,
	}

	pEnc, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}

	return pEnc
}

func listInitiatorRules(mp *map[string]interface{}) []byte {
	m := *mp

	a := m["SourceAccount"].(string)
	i := m["Initiator"].(string)

	payload := PayloadListInitiatorRules{
		SourceAccount: a,
		Initiator:     i,
	}

	pEnc, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}

	return pEnc
}
func addInitiatorToGroup(mp *map[string]interface{}) []byte {
	m := *mp

	a := m["SourceAccount"].(string)
	i := m["Initiator"].(string)
	g := m["Group"].(string)

	payload := PayloadAddInitiatorToGroup{
		SourceAccount: a,
		Initiator:     i,
		Group:         g,
	}

	pEnc, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}

	return pEnc
}

func removeInitiatorFromGroup(mp *map[string]interface{}) []byte {
	m := *mp

	a := m["SourceAccount"].(string)
	i := m["Initiator"].(string)
	g := m["Group"].(string)

	payload := PayloadRemoveInitiatorFromGroup{
		SourceAccount: a,
		Initiator:     i,
		Group:         g,
	}

	pEnc, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}

	return pEnc
}

func listInitiatorGroups(mp *map[string]interface{}) []byte {
	m := *mp

	a := m["SourceAccount"].(string)
	i := m["Initiator"].(string)

	payload := PayloadListInitiatorGroups{
		SourceAccount: a,
		Initiator:     i,
	}

	pEnc, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}

	return pEnc
}

func setInitiatorPubKeys(mp *map[string]interface{}) []byte {
	m := *mp

	a := m["SourceAccount"].(string)
	i := m["Initiator"].(string)
	g := m["PubKeys"].(string)

	payload := PayloadSetInitiatorPubKeys{
		SourceAccount: a,
		Initiator:     i,
		PubKeys:       strings.Split(g, ","),
	}

	pEnc, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}

	return pEnc
}

func deleteInitiatorPubKeys(mp *map[string]interface{}) []byte {
	m := *mp

	a := m["SourceAccount"].(string)
	i := m["Initiator"].(string)
	g := m["PubKeys"].(string)

	payload := PayloadDeleteInitiatorPubKeys{
		SourceAccount: a,
		Initiator:     i,
		PubKeys:       strings.Split(g, ","),
	}

	pEnc, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}

	return pEnc
}

func listInitiatorPubKeys(mp *map[string]interface{}) []byte {
	m := *mp

	a := m["SourceAccount"].(string)
	i := m["Initiator"].(string)

	payload := PayloadListInitiatorPubKeys{
		SourceAccount: a,
		Initiator:     i,
	}

	pEnc, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}

	return pEnc
}

func queryAuth(mp *map[string]interface{}) []byte {
	m := *mp

	a := m["SourceAccount"].(string)
	i := m["Initiator"].(string)
	r := m["Recipient"].(string)
	c := m["Action"].(string)
	n := m["Amount"].(float64)
	d := m["DestAccount"].(string)

	payload := PayloadQueryAuth{
		SourceAccount: a,
		Initiator:     i,
		Recipient:     r,
		Action:        c,
		Amount:        n,
		DestAccount:   d,
	}

	pEnc, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}

	return pEnc
}

func closePendingTx(mp *map[string]interface{}) []byte {
	m := *mp

	a := m["SourceAccount"].(string)
	i := m["Initiator"].(string)
	t := m["TransactionID"].(string)
	k := m["InitiatorKey"].([]byte)

	payload := PayloadClosePendingTx{
		SourceAccount: a,
		Initiator:     i,
		InitiatorKey:  k,
		TransactionID: t,
	}

	pEnc, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}

	return pEnc
}

func addSigTx(mp *map[string]interface{}) []byte {
	m := *mp

	a := m["SourceAccount"].(string)
	i := m["Initiator"].(string)
	t := m["TransactionID"].(string)
	s := m["Signature"].(string)
	k := m["PubKeys"].(string)

	if len(strings.Split(k, ",")) != 1 {
		panic("only one key can be passed when adding a signature to a pending transaction")
	}

	pubKey, err := hex.DecodeString(k)
	if err != nil {
		panic(err)
	}

	sig, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}

	payload := PayloadAddSigTx{
		SourceAccount: a,
		Initiator:     i,
		TransactionID: t,
		Signature:     sig,
		PubKey:        pubKey,
	}

	pEnc, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}

	return pEnc
}

func listPendingTx(mp *map[string]interface{}) []byte {
	m := *mp

	a := m["SourceAccount"].(string)
	i := m["Initiator"].(string)

	payload := PayloadListPendingTx{
		SourceAccount: a,
		Initiator:     i,
	}

	pEnc, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}

	return pEnc
}

func setRecipient(mp *map[string]interface{}) []byte {
	m := *mp

	a := m["SourceAccount"].(string)
	r := m["Recipient"].(string)
	d := m["DestAccount"].(string)

	payload := PayloadSetRecipient{
		SourceAccount: a,
		Recipient:     r,
		DestAccount:   d,
	}

	pEnc, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}

	return pEnc
}

func removeRecipient(mp *map[string]interface{}) []byte {
	m := *mp

	a := m["SourceAccount"].(string)
	r := m["Recipient"].(string)
	d := m["DestAccount"].(string)

	payload := PayloadRemoveRecipient{
		SourceAccount: a,
		Recipient:     r,
		DestAccount:   d,
	}

	pEnc, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}

	return pEnc
}

func listRecipient(mp *map[string]interface{}) []byte {
	m := *mp

	a := m["SourceAccount"].(string)
	r := m["Recipient"].(string)

	payload := PayloadListRecipient{
		SourceAccount: a,
		Recipient:     r,
	}

	pEnc, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}

	return pEnc
}

func setAccountLevelRule(mp *map[string]interface{}) []byte {
	m := *mp

	a := m["SourceAccount"].(string)
	r := m["Rule"].(string)

	payload := PayloadSetInitiatorRule{
		SourceAccount: a,
		Initiator:     DefaultGroupName,
		Rule:          r,
	}

	pEnc, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}

	return pEnc
}

func deleteAccountLevelRule(mp *map[string]interface{}) []byte {
	m := *mp

	a := m["SourceAccount"].(string)
	r := m["RuleHash"].(string)

	payload := PayloadDeleteInitiatorRule{
		SourceAccount: a,
		Initiator:     DefaultGroupName,
		RuleHash:      r,
	}

	pEnc, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}

	return pEnc
}

func listAccountLevelRules(mp *map[string]interface{}) []byte {
	m := *mp

	a := m["SourceAccount"].(string)

	payload := PayloadListInitiatorRules{
		SourceAccount: a,
		Initiator:     DefaultGroupName,
	}

	pEnc, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}

	return pEnc
}

// SignedPayload satisfies an introspection need when unmarshalling payloads
type SignedPayload struct {
	SourceAccount string
	Type          string
	SignerPubKey  []byte
	Signature     []byte
	Payload       []byte `json:"payload"`
}

// PayloadSetInitiatorRule for setting new rules
type PayloadSetInitiatorRule struct {
	SourceAccount string
	Initiator     string `json:"initiator"` // because every rule is attached to an initiator which can be a group. we don't need to know whether Initiator is a Transactor{} or a group. Why? the logic for setting the rule is the same.
	Rule          string `json:"rule"`
}

// PayloadDeleteInitiatorRule for deleting existing rules. Users are assumed to choose to delete a rule after listing rules, in which hashes also appear
type PayloadDeleteInitiatorRule struct {
	SourceAccount string
	Initiator     string
	RuleHash      string
}

// PayloadListInitiatorRules for rule queries, i.e., which rules apply to initiator
type PayloadListInitiatorRules struct {
	SourceAccount string
	Initiator     string
}

// PayloadListInitiatorGroups for rule queries, i.e., which rules apply to initiator
type PayloadListInitiatorGroups struct {
	SourceAccount string
	Initiator     string
}

// PayloadListInitiatorPubKeys for rule queries, i.e., which rules apply to initiator
type PayloadListInitiatorPubKeys struct {
	SourceAccount string
	Initiator     string // list this initiator's pub keys
}

// PayloadAddInitiatorToGroup for setting group membership on an initiator
type PayloadAddInitiatorToGroup struct {
	SourceAccount string
	Initiator     string
	Group         string
}

// PayloadRemoveInitiatorFromGroup for setting group membership on an initiator
type PayloadRemoveInitiatorFromGroup struct {
	SourceAccount string
	Initiator     string
	Group         string
}

// PayloadSetInitiatorPubKeys for attaching pub keys to Initiator. Multiple keys because one for mobile, one for desktop, etc.
type PayloadSetInitiatorPubKeys struct {
	SourceAccount string
	Initiator     string   `json:"initiator"` // the pub keys belong to this initiator
	PubKeys       []string `json:"pub_keys"`
}

// PayloadDeleteInitiatorPubKeys for deleting  pub keys to Initiator
type PayloadDeleteInitiatorPubKeys struct {
	SourceAccount string
	Initiator     string   `json:"initiator"` // the pub keys belong to this initiator
	PubKeys       []string `json:"handle"`    // Would be nice to replace this with a handle, like "mobile" or "desktop", etc., instead of explicitly passing the horrid public keys
}

// PayloadQueryAuth for querying about acceptance/rejection of transactions (on accounts; not blockchain transactions)
type PayloadQueryAuth struct {
	SourceAccount string
	Initiator     string
	Recipient     string
	Action        string
	Amount        float64 `json:"amount"`
	DestAccount   string
}

// PayloadClosePendingTx for closing pending tx so the user can cancel pending tx
type PayloadClosePendingTx struct {
	SourceAccount string
	TransactionID string
	Initiator     string // this is the intiator who wants to close the pending transaction
	InitiatorKey  []byte
}

// PayloadAddSigTx to add a signature to a pending tx awaiting multi sigs
type PayloadAddSigTx struct {
	SourceAccount string
	TransactionID string
	Signature     []byte
	PubKey        []byte // Initiator's public key needed to verify Signature
	Initiator     string // this is the initiator who wants to add its signature
}

// PayloadListPendingTx list all pending transactions that need this initiator's sig
type PayloadListPendingTx struct {
	SourceAccount string
	Initiator     string
}

// PayloadSetRecipient for setting new rules (tie recipient to account for now)
type PayloadSetRecipient struct {
	SourceAccount string
	Recipient     string
	DestAccount   string
}

// PayloadRemoveRecipient for setting new rules
type PayloadRemoveRecipient struct {
	SourceAccount string
	Recipient     string
	DestAccount   string
}

// PayloadListRecipient list all rules tied to this recipient
type PayloadListRecipient struct {
	SourceAccount string
	Recipient     string
}

// ResponseGateway meant to be sent back through AWS API gateway
type ResponseGateway struct {
	Response []string
	Error    string
}
