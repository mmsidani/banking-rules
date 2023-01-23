package core

import (
	"reflect"
)

// PayloadRegistry map strings to payload types
var PayloadRegistry = map[string]reflect.Type{
	"set_initiator_rule":          reflect.TypeOf(&PayloadSetInitiatorRule{}),
	"delete_initiator_rule":       reflect.TypeOf(&PayloadDeleteInitiatorRule{}),
	"list_initiator_rules":        reflect.TypeOf(&PayloadListInitiatorRules{}),
	"add_initiator_to_group":      reflect.TypeOf(&PayloadAddInitiatorToGroup{}),
	"remove_initiator_from_group": reflect.TypeOf(&PayloadRemoveInitiatorFromGroup{}),
	"list_initiator_groups":       reflect.TypeOf(&PayloadListInitiatorGroups{}),
	"set_initiator_pub_keys":      reflect.TypeOf(&PayloadSetInitiatorPubKeys{}),
	"delete_initiator_pub_keys":   reflect.TypeOf(&PayloadDeleteInitiatorPubKeys{}),
	"list_initiator_pub_keys":     reflect.TypeOf(&PayloadListInitiatorPubKeys{}),
	"query_auth":                  reflect.TypeOf(&PayloadQueryAuth{}),
	"set_pending_tx":              reflect.TypeOf(&PayloadSetPendingTx{}),
	"close_pending_tx":            reflect.TypeOf(&PayloadClosePendingTx{}),
	"add_sig_tx":                  reflect.TypeOf(&PayloadAddSigTx{}),
	"list_pending_tx":             reflect.TypeOf(&PayloadListPendingTx{}),
	"set_recipient":               reflect.TypeOf(&PayloadSetRecipient{}),
	"remove_recipient":            reflect.TypeOf(&PayloadRemoveRecipient{}),
	"list_recipient":              reflect.TypeOf(&PayloadListRecipient{}),
	"set_account_level_rule":      reflect.TypeOf(&PayloadSetInitiatorRule{}),
	"delete_account_level_rule":   reflect.TypeOf(&PayloadDeleteInitiatorRule{}),
	"list_account_level_rules":    reflect.TypeOf(&PayloadListInitiatorRules{}),
}

// FamilyName returns family name from source account and payload type
func FamilyName(sourceAccount, payloadType string) string {
	return familyName(sourceAccount, typeToPermissionTag[payloadType])
}

func familyName(sourceAccount, permissionTag string) string {
	return sourceAccount + permissionTag
}
