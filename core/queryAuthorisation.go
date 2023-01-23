package core

import (
	"encoding/hex"
	"encoding/json"

	c "../common"
	"github.com/hyperledger/sawtooth-sdk-go/protobuf/transaction_pb2"
)

// PayloadQueryAuth for querying the system about allowing a banking transaction to go through
type PayloadQueryAuth c.PayloadQueryAuth

// Handle to handle authorisation queries payloads
func (*PayloadQueryAuth) Handle(pl []byte) map[string]interface{} {
	var p PayloadQueryAuth
	err := json.Unmarshal(pl, &p)
	if err != nil {
		panic(err)
	}

	m := c.Struct2Map(&p)

	ret := queryRules(p.SourceAccount, p.Initiator, m)
	if ret["action"] == "allow" || ret["action"] == "deny" {
		return ret
	}

	// we got here, therefore ret["action"] == "pending"
	ptx := createSetPendingTx(p.SourceAccount, p.Initiator, pl, ret)

	// SubmitTx() returns linkToStatus after it has polled until the batch was processed. so, if we're here, we're really done and we have more important things to return, so we ignore linkToStatus
	_ = SubmitTx(ptx)

	return ret
}

func queryRules(sourceAccount, initiator string, m map[string]interface{}) map[string]interface{} {
	// root address of initiator rules, groups, etc.
	initiatorRootAddress := initiatorRootStateAddress(sourceAccount)
	// individual rules
	rulesAddress := initiatorWildCardRules(initiatorRootAddress, initiator)
	_, rules := SubmitStateReq(rulesAddress)

	// group rules
	groupsAddress := initiatorWildCardGroups(initiatorRootAddress, initiator)
	_, groups := SubmitStateReq(groupsAddress)
	// now add the default group. account level rules are assigned to this group
	groups = append(groups, []byte(c.DefaultGroupName))
	gRules := make([][]byte, 0)
	for _, group := range groups {
		g := string(group)
		// Note a group has its rules stored in the state under the same address structure as an individual 'initiator'. essentially a rule for a group=group_name is a rule for initiator=group_name
		groupRulesAddress := initiatorWildCardRules(initiatorRootAddress, g)
		_, r := SubmitStateReq(groupRulesAddress)
		gRules = append(gRules, r...)
	}

	// TODO allow unless a rule rejects it. other possibility is: deny unless rule allows it. make it settable??
	grs := unmarshalRules(gRules)
	irs := unmarshalRules(rules)
	accountRules := SortConflicts(map[string][]ARule{
		"gen":  grs,
		"spec": irs,
	})

	var violatedRules []string     // list of violated rules prepended with their hashes
	var authorisedSigners []string // list of lists of authorised signers. each entry is a comma-separated list
	var minNumberOfSigners []int
	for _, r := range accountRules {
		ev := r.Evaluate(m)
		if ev != "nil" { // rule evaluates to nil when no action required
			// rule was triggered
			if ev == "deny" {
				violatedRules = append(violatedRules, r.RuleHash+":"+r.Rule)
				// note we don't return after the first "deny". we want to report all the rules that trigger "deny"
			} else if len(violatedRules) == 0 {
				// we get here if a signoff is required for instance. but if a rule has triggered "deny" there's no point. that's why we check for length of violatedRules array
				arInt := ev.([]interface{})
				// govaluate doesn't see int's, only float64's. hence the acrobatics here
				minNumberOfSigners = append(minNumberOfSigners, int(arInt[0].(float64)))
				authorisedSigners = append(authorisedSigners, arInt[1].(string))

			}
		}
	}

	if len(violatedRules) != 0 {
		return map[string]interface{}{"action": "deny", "violated_rules": violatedRules}
	}

	if len(authorisedSigners) != 0 {
		return map[string]interface{}{"action": "pending", "authorised_sigs": authorisedSigners, "min_required_sigs": minNumberOfSigners}
	}

	return map[string]interface{}{"action": "allow"}
}

// Note: this lives here and not in payloadPending.go because the keys of the sigs argument which are only known here
func createSetPendingTx(sourceAccount, initiator string, pl []byte, sigs map[string]interface{}) *transaction_pb2.Transaction {
	// first create PayloadSetPendingTx and set unique id to signature of the query auth payload
	bankPubKey, signer := GetBankAuthTools()
	uid := hex.EncodeToString(signer.Sign(pl)[:]) // this is just used here as a unique identifier
	pendingTxPayload := PayloadSetPendingTx{
		SourceAccount:   sourceAccount,
		BankTransaction: pl,
		AuthorisedSigs:  sigs["authorised_sigs"].([]string),
		RequiredMinSigs: sigs["min_required_sigs"].([]int),
		TransactionID:   formatPendingTxUID(uid),
		Initiator:       initiator,
	}

	// now create SignedPayload to wrap in transaction
	payloadEnc, err := json.Marshal(pendingTxPayload)
	if err != nil {
		panic(err)
	}
	signature := signer.Sign(payloadEnc)
	signedPayload := c.SignedPayload{
		Type:         "set_pending_tx",
		SignerPubKey: bankPubKey.AsBytes(),
		Signature:    signature,
		Payload:      payloadEnc,
	}

	pendingRoot := pendingTxStateRootAddress(sourceAccount)
	pendingTxSigs := pendingTxSigs(pendingRoot, uid)
	pendingTxTx := pendingTxTx(pendingRoot, uid)
	initiatorAddress := pendingTxInitiator(pendingRoot, uid)
	outputs := []string{pendingTxSigs, pendingTxTx, initiatorAddress}
	inputs := outputs // not sure why we need to have anything in the inputs
	dependencies := []string{}
	// Note: family names for pending transactions do not take permission tags
	familyName := pendingTxPayload.SourceAccount

	return CreateTransaction(&signedPayload, familyName, inputs, outputs, dependencies)
}
