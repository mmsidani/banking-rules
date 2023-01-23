package core

import (
	"encoding/json"
	"errors"

	c "../common"
	"github.com/hyperledger/sawtooth-sdk-go/processor"
	"github.com/hyperledger/sawtooth-sdk-go/protobuf/transaction_pb2"
)

// PayloadSetRecipient methods implementation
type PayloadSetRecipient c.PayloadSetRecipient

// PayloadRemoveRecipient methods implementation
type PayloadRemoveRecipient c.PayloadRemoveRecipient

// PayloadListRecipient methods implementation
type PayloadListRecipient c.PayloadListRecipient

const (
	// state addresses of accounts attached to recipients take this "mid-fix"
	recipientNamespace = "02"
	accountsSubspace   = "01"
)

// Apply tie recipient to specific accounts
func (*PayloadSetRecipient) Apply(pl []byte, context *processor.Context) error {
	var p PayloadSetRecipient
	err := json.Unmarshal(pl, &p)
	if err != nil {
		panic(err)
	}

	address := recipientAccount(recipientRootStateAddress(p.SourceAccount), p.Recipient, p.DestAccount)

	a := []byte(p.DestAccount)

	addresses, err := context.SetState(map[string][]byte{
		address: a,
	})
	if err != nil || len(addresses) == 0 {
		// TODO change return type to error code
		return errors.New("error setting recipient")
	}

	return nil
}

// Apply remove account details for recipient
func (*PayloadRemoveRecipient) Apply(pl []byte, context *processor.Context) error {
	var p PayloadRemoveRecipient
	err := json.Unmarshal(pl, &p)
	if err != nil {
		panic(err)
	}

	address := recipientAccount(recipientRootStateAddress(p.SourceAccount), p.Recipient, p.DestAccount)

	addresses, err := context.DeleteState([]string{address})
	if err != nil || len(addresses) == 0 {
		// TODO change return type to error code
		return errors.New("error setting recipient")
	}

	return nil
}

// Handle list all state information about this recipient
func (*PayloadListRecipient) Handle(pl []byte) map[string]interface{} {
	var p PayloadListRecipient
	err := json.Unmarshal(pl, &p)
	if err != nil {
		panic(err)
	}

	// TODO this assumes only info we have about recipients is accounts. when we have other information we will change the wild card address to recipientWildCard()
	address := recipientWildCardAccounts(recipientRootStateAddress(p.SourceAccount), p.Recipient)
	_, rules := SubmitStateReq(address)

	rl := make([]string, len(rules))
	for i, r := range rules {
		rl[i] = p.Recipient + ":" + string(r)
	}

	return map[string]interface{}{
		"accounts": rl,
	}

}

// WrapInTx wrap SignedPayload with PayloadSetRecipient payload in a sawtooth transaction
func (*PayloadSetRecipient) WrapInTx(pl *c.SignedPayload) *transaction_pb2.Transaction {
	ok := VerifySignature(pl.Payload, pl.Signature, pl.SignerPubKey)
	if !ok {
		panic("Invalid signature for set recipient transaction ")
	}

	var p PayloadSetRecipient
	err := json.Unmarshal(pl.Payload, &p)
	if err != nil {
		panic(err)
	}

	return wrapRecipientPlInTx(pl, p.SourceAccount, p.Recipient, p.DestAccount, pl.SignerPubKey)
}

// WrapInTx wrap SignedPayload with PayloadRemoveRecipient payload in a sawtooth transaction
func (*PayloadRemoveRecipient) WrapInTx(pl *c.SignedPayload) *transaction_pb2.Transaction {
	ok := VerifySignature(pl.Payload, pl.Signature, pl.SignerPubKey)
	if !ok {
		panic("Invalid signature for set recipient transaction ")
	}

	var p PayloadRemoveRecipient
	err := json.Unmarshal(pl.Payload, &p)
	if err != nil {
		panic(err)
	}

	return wrapRecipientPlInTx(pl, p.SourceAccount, p.Recipient, p.DestAccount, pl.SignerPubKey)
}

func wrapRecipientPlInTx(pl *c.SignedPayload, sourceAccount, recipient, destAccount string, pubKey []byte) *transaction_pb2.Transaction {
	root := recipientRootStateAddress(sourceAccount)
	outputs := []string{recipientAccount(root, recipient, destAccount)}
	inputs := outputs
	dependencies := []string{}
	fn := familyName(sourceAccount, RecipientPermissionTag)

	ok := VerifyPermission(fn, pubKey)
	if !ok {
		panic("signer of recipient transaction setting is not authorised")
	}

	return CreateTransaction(pl, fn, inputs, outputs, dependencies)
}

type recipientRootAddressType string

func recipientRootStateAddress(sourceAccount string) recipientRootAddressType {
	return recipientRootAddressType(Namespace(familyName(sourceAccount, RecipientPermissionTag)) + recipientNamespace)
}

func recipientWildCard(root recipientRootAddressType, recipient string) string {
	return string(root) + HexdigestStr(recipient)[:actorLength]
}

func recipientWildCardAccounts(root recipientRootAddressType, recipient string) string {
	return recipientWildCard(root, recipient) + accountsSubspace
}

func recipientAccount(root recipientRootAddressType, recipient, destAccount string) string {
	return CheckLength(recipientWildCardAccounts(root, recipient) + HexdigestStr(destAccount)[:fieldLength])
}
