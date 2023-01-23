package core

// TODO TODO deprecated about to be deleted

// // // import (
// // // 	"encoding/json"
// // // 	"errors"

// // // 	c "../common"
// // // 	"github.com/hyperledger/sawtooth-sdk-go/processor"
// // // 	"github.com/hyperledger/sawtooth-sdk-go/protobuf/transaction_pb2"
// // // )

// // // // PayloadSetMultiSigRule for setting multi-sig rules
// // // type PayloadSetMultiSigRule c.PayloadSetMultiSigRule

// // // // PayloadDeleteMultiSigRule for deleting multi-sig rules
// // // type PayloadDeleteMultiSigRule c.PayloadDeleteMultiSigRule

// // // // PayloadListMultiSigRules for list all multi signature rules on the account
// // // type PayloadListMultiSigRules c.PayloadListMultiSigRules

// // // const (
// // // 	multiSigRulesNamespace = "10"
// // // 	fillerSubspace         = "00"
// // // )

// // // // Apply for setting requirements for  multiple signatures
// // // func (*PayloadSetMultiSigRule) Apply(pl []byte, context *processor.Context) error {
// // // 	var p PayloadSetMultiSigRule
// // // 	err := json.Unmarshal(pl, &p)
// // // 	if err != nil {
// // // 		panic(err)
// // // 	}

// // // 	address := multiSigRule(multiSigRootStateAddress(p.SourceAccount), p.Rule)
// // // 	ruleHash := getMultiSigRuleHash(address)
// // // 	rule := p.Rule

// // // 	newRule := NewMultiSigRule(rule, ruleHash)

// // // 	r, err := json.Marshal(newRule)
// // // 	if err != nil {
// // // 		panic(err)
// // // 	}

// // // 	addresses, err := context.SetState(map[string][]byte{
// // // 		address: r,
// // // 	})
// // // 	if err != nil || len(addresses) == 0 {
// // // 		return errors.New("error setting new multi sig rule")
// // // 	}

// // // 	return nil
// // // }

// // // // Apply for deleting multiple signature requirements
// // // func (*PayloadDeleteMultiSigRule) Apply(pl []byte, context *processor.Context) error {
// // // 	var p PayloadDeleteMultiSigRule
// // // 	err := json.Unmarshal(pl, &p)
// // // 	if err != nil {
// // // 		panic(err)
// // // 	}

// // // 	address := multiSigRuleHash(multiSigRootStateAddress(p.SourceAccount), p.RuleHash)

// // // 	addresses, err := context.DeleteState([]string{address})
// // // 	if err != nil || len(addresses) == 0 {
// // // 		return errors.New("error deleting multi sig rule")
// // // 	}

// // // 	return nil
// // // }

// // // // Handle list all multi sig rules set on account
// // // func (*PayloadListMultiSigRules) Handle(pl []byte) map[string]interface{} {
// // // 	var p PayloadListMultiSigRules
// // // 	err := json.Unmarshal(pl, &p)
// // // 	if err != nil {
// // // 		panic(err)
// // // 	}

// // // 	ruleHashes, rules := extractMultiSigRules(&p)

// // // 	return map[string]interface{}{
// // // 		"multi_signature_rules":        rules,
// // // 		"multi_signature_rules_hashes": ruleHashes,
// // // 	}
// // // }

// // // // WrapInTx wrap set multi sig rule payload to submit to validator
// // // func (*PayloadSetMultiSigRule) WrapInTx(pl *c.SignedPayload) *transaction_pb2.Transaction {
// // // 	ok := VerifySignature(pl.Payload, pl.Signature, pl.SignerPubKey)
// // // 	if !ok {
// // // 		panic("Invalid signature for set multi sig rule transaction ")
// // // 	}

// // // 	var p PayloadSetMultiSigRule
// // // 	err := json.Unmarshal(pl.Payload, &p)
// // // 	if err != nil {
// // // 		panic(err)
// // // 	}

// // // 	outputs := []string{multiSigRule(multiSigRootStateAddress(p.SourceAccount), p.Rule)}
// // // 	inputs := outputs
// // // 	dependencies := []string{}
// // // 	fn := familyName(p.SourceAccount, MultiSigPermissionTag)
// // // 	ok = VerifyPermission(fn, pl.SignerPubKey)
// // // 	if !ok {
// // // 		panic("signer of multi sig transaction is not authorised")
// // // 	}

// // // 	return CreateTransaction(pl, fn, inputs, outputs, dependencies)
// // // }

// // // // WrapInTx wrap delete multi sig rule payload to submit to validator
// // // func (*PayloadDeleteMultiSigRule) WrapInTx(pl *c.SignedPayload) *transaction_pb2.Transaction {
// // // 	ok := VerifySignature(pl.Payload, pl.Signature, pl.SignerPubKey)
// // // 	if !ok {
// // // 		panic("Invalid signature for delete multi sig rule transaction ")
// // // 	}

// // // 	var p PayloadDeleteMultiSigRule
// // // 	err := json.Unmarshal(pl.Payload, &p)
// // // 	if err != nil {
// // // 		panic(err)
// // // 	}

// // // 	outputs := []string{multiSigRuleHash(multiSigRootStateAddress(p.SourceAccount), p.RuleHash)}
// // // 	inputs := outputs
// // // 	dependencies := []string{}
// // // 	fn := familyName(p.SourceAccount, MultiSigPermissionTag)
// // // 	ok = VerifyPermission(fn, pl.SignerPubKey)
// // // 	if !ok {
// // // 		panic("signer of multi sig transaction is not authorised")
// // // 	}

// // // 	return CreateTransaction(pl, fn, inputs, outputs, dependencies)
// // // }

// // // type multiSigRootAddressType string

// // // func multiSigRootStateAddress(sourceAccount string) multiSigRootAddressType {
// // // 	return multiSigRootAddressType(Namespace(familyName(sourceAccount, MultiSigPermissionTag)) + multiSigRulesNamespace)
// // // }

// // // // all multisig rules go under here
// // // func multiSigRulesWildCard(root multiSigRootAddressType) string {
// // // 	return string(root) + fillerSubspace
// // // }

// // // // address of specific multisig rule. Note: I did whatever to meet address length requirement.
// // // func multiSigRule(root multiSigRootAddressType, rule string) string {
// // // 	ruleHash := HexdigestStr(rule)
// // // 	return CheckLength(multiSigRulesWildCard(root) + ruleHash[:actorLength+fieldLength])
// // // }

// // // // reconstitute the multisig rule address from ruleHash. needed to delete multisig rules, e.g., since we only pass the hash in the payload
// // // func multiSigRuleHash(root multiSigRootAddressType, ruleHash string) string {
// // // 	return CheckLength(multiSigRulesWildCard(root) + ruleHash)
// // // }

// // // // get rule hash that was used in building rule address. Note: this results in a very different rule hash length from regular rules. should it matter?
// // // func getMultiSigRuleHash(address string) string {
// // // 	return address[(AddressLength - actorLength - fieldLength):]
// // // }

// // // func extractMultiSigRules(pl *PayloadListMultiSigRules) ([]string, []string) {
// // // 	address := multiSigRulesWildCard(multiSigRootStateAddress(pl.SourceAccount))
// // // 	_, rules := SubmitStateReq(address)

// // // 	multiSigRules := unmarshalRules(rules)
// // // 	return tabulateRules(multiSigRules)
// // // }
