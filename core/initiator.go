package core

import (
	"encoding/json"
	"errors"

	c "../common"
	"github.com/hyperledger/sawtooth-sdk-go/processor"
	"github.com/hyperledger/sawtooth-sdk-go/protobuf/transaction_pb2"
)

// Note: here and elsewhere in this package, the state address calculation logic dictated the grouping of different payloads in one file

// PayloadSetInitiatorRule setting rules that apply to individual intiator or group
type PayloadSetInitiatorRule c.PayloadSetInitiatorRule

// PayloadDeleteInitiatorRule delete rule that applies to individual intiator or group
type PayloadDeleteInitiatorRule c.PayloadDeleteInitiatorRule

// PayloadListInitiatorRules list rules that apply to individual intiator or group
type PayloadListInitiatorRules c.PayloadListInitiatorRules

// PayloadAddInitiatorToGroup assign individual initiator or group to a (larger) group
type PayloadAddInitiatorToGroup c.PayloadAddInitiatorToGroup

// PayloadRemoveInitiatorFromGroup remove initiator from group
type PayloadRemoveInitiatorFromGroup c.PayloadRemoveInitiatorFromGroup

// PayloadListInitiatorGroups list all groups that this initiator belongs to
type PayloadListInitiatorGroups c.PayloadListInitiatorGroups

// PayloadSetInitiatorPubKeys assign public key(s) to (individual) initiator
type PayloadSetInitiatorPubKeys c.PayloadSetInitiatorPubKeys

// PayloadDeleteInitiatorPubKeys remove public key that was attached to initiator
type PayloadDeleteInitiatorPubKeys c.PayloadDeleteInitiatorPubKeys

// PayloadListInitiatorPubKeys list all keys that were assigned to this (individual) initiator
type PayloadListInitiatorPubKeys c.PayloadListInitiatorPubKeys

const (
	initiatorNamespace = "01"
	rulesSubspace      = "01"
	groupsSubspace     = "02"
	pubKeysSubspace    = "03"
)

// Apply applier for setting new account rules
func (*PayloadSetInitiatorRule) Apply(pl []byte, context *processor.Context) error {
	var p PayloadSetInitiatorRule
	err := json.Unmarshal(pl, &p)
	if err != nil {
		panic(err)
	}

	address := initiatorRule(initiatorRootStateAddress(p.SourceAccount), p.Initiator, p.Rule)
	ruleHash := getRuleHash(address)
	rule := p.Rule

	newRule := NewRule(rule, ruleHash)

	r, err := json.Marshal(newRule)
	if err != nil {
		panic(err)
	}

	addresses, err := context.SetState(map[string][]byte{
		address: r,
	})
	if err != nil || len(addresses) == 0 {
		return errors.New("error setting new rule")
	}

	return nil
}

// Apply applier for deleting new account rules
func (*PayloadDeleteInitiatorRule) Apply(pl []byte, context *processor.Context) error {
	var p PayloadDeleteInitiatorRule
	err := json.Unmarshal(pl, &p)
	if err != nil {
		panic(err)
	}

	address := initiatorRuleHash(initiatorRootStateAddress(p.SourceAccount), p.Initiator, p.RuleHash)

	addresses, err := context.DeleteState([]string{address})
	if err != nil || len(addresses) == 0 {
		return errors.New("error deleting rule " + p.RuleHash)
	}

	return nil
}

// Apply add a transactor to a group of transactors
func (*PayloadAddInitiatorToGroup) Apply(pl []byte, context *processor.Context) error {
	var p PayloadAddInitiatorToGroup
	err := json.Unmarshal(pl, &p)
	if err != nil {
		panic(err)
	}

	address := initiatorGroup(initiatorRootStateAddress(p.SourceAccount), p.Initiator, p.Group)

	g := []byte(p.Group)

	addresses, err := context.SetState(map[string][]byte{
		address: g,
	})
	if err != nil || len(addresses) == 0 {
		return errors.New("error adding group")
	}

	return nil
}

// Apply remove a transactor from a group of transactors
func (*PayloadRemoveInitiatorFromGroup) Apply(pl []byte, context *processor.Context) error {
	var p PayloadRemoveInitiatorFromGroup
	err := json.Unmarshal(pl, &p)
	if err != nil {
		panic(err)
	}

	address := initiatorGroup(initiatorRootStateAddress(p.SourceAccount), p.Initiator, p.Group)

	addresses, err := context.DeleteState([]string{address})
	if err != nil || len(addresses) == 0 {
		return errors.New("error removing group")
	}

	return nil
}

// Apply set the public keys for a given transactor, typically one for every channel
func (*PayloadSetInitiatorPubKeys) Apply(pl []byte, context *processor.Context) error {
	var p PayloadSetInitiatorPubKeys
	err := json.Unmarshal(pl, &p)
	if err != nil {
		panic(err)
	}

	pubKeys := p.PubKeys
	pkEnc, err := json.Marshal(pubKeys)
	if err != nil {
		panic(err)
	}

	address := initiatorPubKeys(initiatorRootStateAddress(p.SourceAccount), p.Initiator)
	m := map[string][]byte{
		address: pkEnc,
	}

	addresses, err := context.SetState(m)
	if err != nil || len(addresses) == 0 {
		return errors.New("error setting pub keys")
	}

	return nil
}

// Apply delete public keys for transactor
func (*PayloadDeleteInitiatorPubKeys) Apply(pl []byte, context *processor.Context) error {
	var p PayloadDeleteInitiatorPubKeys
	err := json.Unmarshal(pl, &p)
	if err != nil {
		panic(err)
	}

	pubKeysAddress := initiatorPubKeys(initiatorRootStateAddress(p.SourceAccount), p.Initiator)

	m, err := context.GetState([]string{pubKeysAddress})
	if err != nil {
		panic(err)
	}
	var pubKeys []string
	err = json.Unmarshal(m[pubKeysAddress], &pubKeys)
	if err != nil {
		panic(err)
	}

	var remainingKeys = make([]string, 0)
	for _, storedKey := range pubKeys {
		found := false
		for _, inKey := range p.PubKeys {
			if storedKey == inKey {
				found = true
				break
			}
		}
		if !found {
			remainingKeys = append(remainingKeys, storedKey)
		}
	}

	if len(remainingKeys) == 0 {
		// no keys remain. we delete the data in the state
		addresses := []string{pubKeysAddress}
		delAddresses, err := context.DeleteState(addresses)
		if err != nil || len(delAddresses) != len(addresses) {
			// TODO change return type to error code
			return errors.New("error deleting pub key")
		}
	} else {
		enc, err := json.Marshal(remainingKeys)
		if err != nil {
			panic(err)
		}

		// set the state to the remaining keys
		addresses, err := context.SetState(map[string][]byte{pubKeysAddress: enc})
		if err != nil || len(addresses) == 0 {
			panic(err)
		}
	}
	return nil
}

// Handle for listing of initiator and recipient specific rules
func (*PayloadListInitiatorRules) Handle(pl []byte) map[string]interface{} {
	var p PayloadListInitiatorRules
	err := json.Unmarshal(pl, &p)
	if err != nil {
		panic(err)
	}

	hashes, rules := extractInitiatorRules(&p)

	return map[string]interface{}{
		"rules":     rules,
		"hashes":    hashes,
		"initiator": p.Initiator,
	}
}

// Handle for listing of initiator and recipient specific rules
func (*PayloadListInitiatorGroups) Handle(pl []byte) map[string]interface{} {
	var p PayloadListInitiatorGroups
	err := json.Unmarshal(pl, &p)
	if err != nil {
		panic(err)
	}

	address := initiatorWildCardGroups(initiatorRootStateAddress(p.SourceAccount), p.Initiator)
	_, groupsB := SubmitStateReq(address)

	ret := make([]string, len(groupsB))
	for i, group := range groupsB {
		ret[i] = string(group)
	}

	return map[string]interface{}{
		"groups":    ret,
		"initiator": p.Initiator,
	}
}

// Handle for listing of initiator and recipient specific rules
func (*PayloadListInitiatorPubKeys) Handle(pl []byte) map[string]interface{} {
	var p PayloadListInitiatorPubKeys
	err := json.Unmarshal(pl, &p)
	if err != nil {
		panic(err)
	}

	address := initiatorPubKeys(initiatorRootStateAddress(p.SourceAccount), p.Initiator)
	_, pubKeys := SubmitStateReq(address)

	var keys []string
	err = json.Unmarshal(pubKeys[0], &keys)
	if err != nil {
		panic(err)
	}

	return map[string]interface{}{
		"pubKeys":   keys,
		"initiator": p.Initiator,
	}
}

// WrapInTx SignedPayload with PayloadSetInitiatorRule to submit to validator
func (*PayloadSetInitiatorRule) WrapInTx(pl *c.SignedPayload) *transaction_pb2.Transaction {
	ok := VerifySignature(pl.Payload, pl.Signature, pl.SignerPubKey)
	if !ok {
		panic("Invalid signature for set initiator rule transaction ")
	}

	var p PayloadSetInitiatorRule
	err := json.Unmarshal(pl.Payload, &p)
	if err != nil {
		panic(err)
	}

	outputs := []string{initiatorRule(initiatorRootStateAddress(p.SourceAccount), p.Initiator, p.Rule)}
	inputs := outputs
	dependencies := []string{}
	fn := familyName(p.SourceAccount, InitiatorPermissionTag)
	ok = VerifyPermission(fn, pl.SignerPubKey)
	if !ok {
		panic("signer of set initiator rule transaction is not authorised")
	}

	return CreateTransaction(pl, fn, inputs, outputs, dependencies)
}

// WrapInTx SignedPayload with PayloadDeleteInitiatorRule to submit to validator
func (*PayloadDeleteInitiatorRule) WrapInTx(pl *c.SignedPayload) *transaction_pb2.Transaction {
	ok := VerifySignature(pl.Payload, pl.Signature, pl.SignerPubKey)
	if !ok {
		panic("Invalid signature for delete initiator rule transaction ")
	}

	var p PayloadDeleteInitiatorRule
	err := json.Unmarshal(pl.Payload, &p)
	if err != nil {
		panic(err)
	}

	outputs := []string{initiatorRuleHash(initiatorRootStateAddress(p.SourceAccount), p.Initiator, p.RuleHash)}
	inputs := outputs
	dependencies := []string{}
	fn := familyName(p.SourceAccount, InitiatorPermissionTag)
	ok = VerifyPermission(fn, pl.SignerPubKey)
	if !ok {
		panic("signer of delete initiator rule transaction is not authorised")
	}

	return CreateTransaction(pl, fn, inputs, outputs, dependencies)
}

// WrapInTx SignedPayload with PayloadAddInitiatorToGroup to submit to validator
func (*PayloadAddInitiatorToGroup) WrapInTx(pl *c.SignedPayload) *transaction_pb2.Transaction {
	ok := VerifySignature(pl.Payload, pl.Signature, pl.SignerPubKey)
	if !ok {
		panic("Invalid signature for add initiator to group transaction ")
	}

	var p PayloadAddInitiatorToGroup
	err := json.Unmarshal(pl.Payload, &p)
	if err != nil {
		panic(err)
	}

	outputs := []string{initiatorGroup(initiatorRootStateAddress(p.SourceAccount), p.Initiator, p.Group)}
	inputs := outputs
	dependencies := []string{}
	fn := familyName(p.SourceAccount, InitiatorPermissionTag)
	ok = VerifyPermission(fn, pl.SignerPubKey)
	if !ok {
		panic("signer of add initiator to group transaction is not authorised")
	}

	return CreateTransaction(pl, fn, inputs, outputs, dependencies)
}

// WrapInTx SignedPayload with PayloadRemoveInitiatorFromGroup to submit to validator
func (*PayloadRemoveInitiatorFromGroup) WrapInTx(pl *c.SignedPayload) *transaction_pb2.Transaction {
	ok := VerifySignature(pl.Payload, pl.Signature, pl.SignerPubKey)
	if !ok {
		panic("Invalid signature for remove initiator from group transaction ")
	}

	var p PayloadRemoveInitiatorFromGroup
	err := json.Unmarshal(pl.Payload, &p)
	if err != nil {
		panic(err)
	}

	outputs := []string{initiatorGroup(initiatorRootStateAddress(p.SourceAccount), p.Initiator, p.Group)}
	inputs := outputs
	dependencies := []string{}
	fn := familyName(p.SourceAccount, InitiatorPermissionTag)
	ok = VerifyPermission(fn, pl.SignerPubKey)
	if !ok {
		panic("signer of remove initiator from group transaction is not authorised")
	}

	return CreateTransaction(pl, fn, inputs, outputs, dependencies)
}

// WrapInTx SignedPayload with PayloadSetInitiatorPubKeys to submit to validator
func (*PayloadSetInitiatorPubKeys) WrapInTx(pl *c.SignedPayload) *transaction_pb2.Transaction {
	ok := VerifySignature(pl.Payload, pl.Signature, pl.SignerPubKey)
	if !ok {
		panic("Invalid signature for set initiator pub keys transaction ")
	}

	var p PayloadSetInitiatorPubKeys
	err := json.Unmarshal(pl.Payload, &p)
	if err != nil {
		panic(err)
	}

	outputs := []string{initiatorPubKeys(initiatorRootStateAddress(p.SourceAccount), p.Initiator)}
	inputs := outputs
	dependencies := []string{}
	fn := familyName(p.SourceAccount, InitiatorPermissionTag)
	ok = VerifyPermission(fn, pl.SignerPubKey)
	if !ok {
		panic("signer of set initiator pub keys transaction is not authorised")
	}

	return CreateTransaction(pl, fn, inputs, outputs, dependencies)
}

// WrapInTx SignedPayload with PayloadDeleteInitiatorPubKeys to submit to validator
func (*PayloadDeleteInitiatorPubKeys) WrapInTx(pl *c.SignedPayload) *transaction_pb2.Transaction {
	ok := VerifySignature(pl.Payload, pl.Signature, pl.SignerPubKey)
	if !ok {
		panic("Invalid signature for delete initiator pub keys transaction ")
	}

	var p PayloadDeleteInitiatorPubKeys
	err := json.Unmarshal(pl.Payload, &p)
	if err != nil {
		panic(err)
	}

	outputs := []string{initiatorPubKeys(initiatorRootStateAddress(p.SourceAccount), p.Initiator)}
	inputs := outputs
	dependencies := []string{}
	fn := familyName(p.SourceAccount, InitiatorPermissionTag)
	ok = VerifyPermission(fn, pl.SignerPubKey)
	if !ok {
		panic("signer of delete initiator pub keys transaction is not authorised")
	}

	return CreateTransaction(pl, fn, inputs, outputs, dependencies)
}

type initiatorRootAddressType string

func initiatorRootStateAddress(sourceAccount string) initiatorRootAddressType {
	return initiatorRootAddressType(Namespace(familyName(sourceAccount, InitiatorPermissionTag)) + initiatorNamespace)
}

// wild card for initiator
func initiatorWildCard(root initiatorRootAddressType, initiator string) string {
	return string(root) + HexdigestStr(initiator)[:actorLength]
}

// wild card for all rule addresses
func initiatorWildCardRules(root initiatorRootAddressType, initiator string) string {
	return initiatorWildCard(root, initiator) + rulesSubspace
}

// rule address
func initiatorRule(root initiatorRootAddressType, initiator, rule string) string {
	return CheckLength(initiatorWildCardRules(root, initiator) + HexdigestStr(rule)[:fieldLength])
}

// recover rule address given its (partial) hash
func initiatorRuleHash(root initiatorRootAddressType, initiator, ruleHash string) string {
	return CheckLength(initiatorWildCardRules(root, initiator) + ruleHash)
}

// get rule hash that was used in building rule address
func getRuleHash(address string) string {
	return address[(AddressLength - fieldLength):]
}

// InitiatorWildCardGroups wild card to retrieve all groups
func initiatorWildCardGroups(root initiatorRootAddressType, initiator string) string {
	return initiatorWildCard(root, initiator) + groupsSubspace
}

// InitiatorGroup this is the address where initiator's belonging to 'group' is recorded
func initiatorGroup(root initiatorRootAddressType, initiator, group string) string {
	return CheckLength(initiatorWildCardGroups(root, initiator) + HexdigestStr(group)[:fieldLength])
}

// address to store pub keys. Note we don't do wild cards for pub keys because we store all of them in one array under one address
func initiatorPubKeys(root initiatorRootAddressType, initiator string) string {
	dummyString := "public keys live here"
	return CheckLength(initiatorWildCard(root, initiator) + pubKeysSubspace + HexdigestStr(dummyString)[:fieldLength])
}

func extractInitiatorRules(pl *PayloadListInitiatorRules) ([]string, []string) {
	address := initiatorWildCardRules(initiatorRootStateAddress(pl.SourceAccount), pl.Initiator)
	_, rules := SubmitStateReq(address)

	accountRules := unmarshalRules(rules)
	return tabulateRules(accountRules)
}
