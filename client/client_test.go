package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"

	"net/http"

	c "../common"
)

func TestSetInitiatorPubKeys(t *testing.T) {
	opts := c.PayloadFields{
		KeysFile:      "/home/majed/.sawtooth/keys/majed",
		RequestType:   "set_initiator_pub_keys",
		SourceAccount: "AB12XF3",
	}

	// set keys for 3 different users
	t.Run("ID12345", func(t *testing.T) {
		// t.Parallel() // to run the subtests in parallel. note that all subtests must call t.Parallel() for the subtests to run in parallel
		opts.Initiator = "ID12345"
		// ID12345 is the same as majed in .sawtooth/keys. here 2 keys from majed.pub and majed_mobile.pub
		opts.PubKeys = "030509c81f5d5e927cd2fbe17ac1e90866e53deec7bba53670afec5e562ef53f9d,02108c5ce04222533516acd943143513470e534e779d5647dd6b58c005c0e7e20b"
		execute(opts)
	})

	t.Run("CD34YG4", func(t *testing.T) {
		opts.Initiator = "CD34YG4"
		opts.PubKeys = "03cf7cfa4a7ce9a5517ab424fc6bc5821db1dc3a2b55483683549947caa5a8a60e"
		execute(opts)
	})

	t.Run("EF56ZH5", func(t *testing.T) {
		opts.Initiator = "EF56ZH5"
		// ID12345 is the same as majed in keys. here 2 keys from majed.pub and majed_mobile.pub
		opts.PubKeys = "02ed6cad6e228f8f08bf82665bfeb158b83e7d4f2c42a04ca9bc335cb12dec715d"
		execute(opts)
	})

}

func TestDeleteInitiatorPubKeys(t *testing.T) {
	opts := c.PayloadFields{
		KeysFile:      "/home/majed/.sawtooth/keys/majed",
		RequestType:   "delete_initiator_pub_keys",
		SourceAccount: "AB12XF3",
		Initiator:     "ID12345",
		PubKeys:       "02108c5ce04222533516acd943143513470e534e779d5647dd6b58c005c0e7e20b", // majed_mobile.pub
	}

	execute(opts)
}

func TestListInitiatorPubKeys(t *testing.T) {
	opts := c.PayloadFields{
		KeysFile:      "/home/majed/.sawtooth/keys/majed",
		RequestType:   "list_initiator_pub_keys",
		SourceAccount: "AB12XF3",
		Initiator:     "ID12345",
	}

	execute(opts)
}

func TestSetInitiatorRule(t *testing.T) {
	opts := c.PayloadFields{
		KeysFile:      "/home/majed/.sawtooth/keys/majed",
		RequestType:   "set_initiator_rule",
		SourceAccount: "AB12XF3",
		Initiator:     "ID12345",
		Rule:          "Amount < 13000",
	}

	execute(opts)
}

func TestListInitiatorRules(t *testing.T) {
	opts := c.PayloadFields{
		KeysFile:      "/home/majed/.sawtooth/keys/majed",
		RequestType:   "list_initiator_rules",
		SourceAccount: "AB12XF3",
		Initiator:     "ID12345",
	}

	execute(opts)
}

func TestDeleteInitiatorRule(t *testing.T) {
	opts := c.PayloadFields{
		KeysFile:      "/home/majed/.sawtooth/keys/majed",
		RequestType:   "delete_initiator_rule",
		SourceAccount: "AB12XF3",
		Initiator:     "ID12345",
		RuleHash:      "3629375f40f4974b0b95",
	}

	execute(opts)
}

func TestSetAccountLevelRule(t *testing.T) {
	opts := c.PayloadFields{
		KeysFile:      "/home/majed/.sawtooth/keys/majed",
		RequestType:   "set_account_level_rule",
		SourceAccount: "AB12XF3",
		Rule:          "Amount > 10000 ? NofM(2, 'ID12345, CD34YG4, EF56ZH5') : 'nil' ",
	}

	execute(opts)
}

func TestListAccountLevelRules(t *testing.T) {
	opts := c.PayloadFields{
		KeysFile:      "/home/majed/.sawtooth/keys/majed",
		RequestType:   "list_account_level_rules",
		SourceAccount: "AB12XF3",
	}

	execute(opts)
}

func TestDeleteAccountLevelRule(t *testing.T) {
	opts := c.PayloadFields{
		KeysFile:      "/home/majed/.sawtooth/keys/majed",
		RequestType:   "delete_account_level_rule",
		SourceAccount: "AB12XF3",
		RuleHash:      "4a2b9ef45c1d328dbe58f83a6f2a4c7a8eb16b956ac23aece1badd49cdb3",
	}

	execute(opts)

}

func TestAddInitiatorToGroup(t *testing.T) {
	opts := c.PayloadFields{
		KeysFile:      "/home/majed/.sawtooth/keys/majed",
		RequestType:   "add_initiator_to_group",
		SourceAccount: "AB12XF3",
		Initiator:     "ID12345",
		Group:         "Partners",
	}

	execute(opts)
}

func TestListInitiatorGroups(t *testing.T) {
	opts := c.PayloadFields{
		KeysFile:      "/home/majed/.sawtooth/keys/majed",
		RequestType:   "list_initiator_groups",
		SourceAccount: "AB12XF3",
		Initiator:     "ID12345",
	}

	execute(opts)

}

func TestRemoveInitiatorFromGroup(t *testing.T) {
	opts := c.PayloadFields{
		KeysFile:      "/home/majed/.sawtooth/keys/majed",
		RequestType:   "remove_initiator_from_group",
		SourceAccount: "AB12XF3",
		Initiator:     "ID12345",
		Group:         "Employees",
	}

	execute(opts)
}

func TestSetRecipient(t *testing.T) {
	opts := c.PayloadFields{
		KeysFile:      "/home/majed/.sawtooth/keys/majed",
		RequestType:   "set_recipient",
		SourceAccount: "AB12XF3",
		Recipient:     "IDR2345",
		DestAccount:   "ZY12ABC",
	}

	execute(opts)
}

func TestListRecipient(t *testing.T) {
	opts := c.PayloadFields{
		KeysFile:      "/home/majed/.sawtooth/keys/majed",
		RequestType:   "list_recipient",
		SourceAccount: "AB12XF3",
		Recipient:     "IDR2345",
	}

	execute(opts)
}

func TestRemoveRecipient(t *testing.T) {
	opts := c.PayloadFields{
		KeysFile:      "/home/majed/.sawtooth/keys/majed",
		RequestType:   "remove_recipient",
		SourceAccount: "AB12XF3",
		Recipient:     "IDR2345",
		DestAccount:   "ZY12ABC",
	}

	execute(opts)
}

func TestQueryAuth(t *testing.T) {
	opts := c.PayloadFields{
		KeysFile:      "/home/majed/.sawtooth/keys/majed",
		RequestType:   "query_auth",
		SourceAccount: "AB12XF3",
		Initiator:     "ID12345",
	}

	// do 2 transactions
	t.Run("11000", func(t *testing.T) {
		opts.Amount = 11000
		execute(opts)
	})

	// t.Run("11500", func(t *testing.T) {
	// 	opts.Amount = 11500
	// 	execute(opts)
	// })
}

func TestListPendingTx(t *testing.T) {
	opts := c.PayloadFields{
		KeysFile:      "/home/majed/.sawtooth/keys/majed",
		RequestType:   "list_pending_tx",
		SourceAccount: "AB12XF3",
		Initiator:     "EF56ZH5",
	}

	execute(opts)
}

func TestClosePendingTx(t *testing.T) {
	opts := c.PayloadFields{
		KeysFile:      "/home/majed/.sawtooth/keys/majed",
		RequestType:   "close_pending_tx",
		SourceAccount: "AB12XF3",
	}

	t.Run("transaction1", func(t *testing.T) {
		opts.Initiator = "ID12345"
		opts.TransactionID = "16be5e25d01c88715cf25ef97dd8688843d568b7516eee9a24d1cf3c2fc3"
		opts.InitiatorKey = "030509c81f5d5e927cd2fbe17ac1e90866e53deec7bba53670afec5e562ef53f9d"
		execute(opts)
	})

	// // this should fail because transaction was not initiated by CD34YG4
	// t.Run("transaction2", func(t *testing.T) {
	// 	opts.Initiator = "CD34YG4"
	// 	opts.TransactionID = ""
	// 	opts.InitiatorKey = "03cf7cfa4a7ce9a5517ab424fc6bc5821db1dc3a2b55483683549947caa5a8a60e"
	// 	execute(opts)
	// })

}

func TestAddSigTx(t *testing.T) {
	// In our use case, the payload of the original transaction that is now pending is displayed to the user who then signs it. So first we reproduce the QueryAuth payload to be signed. Typically, this would be a transaction that we already ran through TestQueryAuth()
	queryAuthPayload := c.PayloadQueryAuth{
		SourceAccount: "AB12XF3",
		Initiator:     "ID12345",
		Amount:        11000,
	}
	qEnc, err := json.Marshal(queryAuthPayload)
	if err != nil {
		panic(err)
	}

	// test signers for which we created key pairs using the 'sawtooth keygen' command
	f := func(keysFile, initiator string) c.PayloadFields {
		signerPrivKey, signerPubKey := c.GetKeysFromFiles(keysFile)
		signer := c.GetSigner(signerPrivKey)
		sig := signer.Sign(qEnc)
		opts := c.PayloadFields{
			KeysFile:      keysFile,
			RequestType:   "add_sig_tx",
			SourceAccount: "AB12XF3",
			Initiator:     initiator,
			Signature:     hex.EncodeToString(sig),
			PubKeys:       signerPubKey.AsHex(),
			TransactionID: "16be5e25d01c88715cf25ef97dd8688843d568b7516eee9a24d1cf3c2fc3",
		}

		return opts
	}

	t.Run("CD34YG4", func(t *testing.T) {
		opts := f("/home/majed/.sawtooth/keys/CD34YG4", "CD34YG4")
		execute(opts)
	})

	// t.Run("EF56ZH5", func(t *testing.T) {
	// 	opts := f("/home/majed/.sawtooth/keys/EF56ZH5", "EF56ZH5")
	// 	execute(opts)
	// })
}

func execute(opts c.PayloadFields) {
	privateKey, publicKey := c.GetKeysFromFiles(opts.KeysFile)
	p := c.CreateSignedPayload(&opts, &privateKey, &publicKey)

	pEnc, err := json.Marshal(*p)
	if err != nil {
		panic(err)
	}

	body := new(bytes.Buffer)
	body.Write(pEnc)

	// TODO change http to https after upgrading to TLS in lambda
	resp, err := http.Post("http://127.0.0.1:8443/", "application/json", body)

	if err != nil {
		panic(err)
	}

	b := new(bytes.Buffer)
	b.ReadFrom(resp.Body)

	fmt.Println(b.String())
}
