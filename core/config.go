package core

// Internal Constants
const (
	EncryptionAlgoName string = "secp256k1"
	FamilyVersion      string = "0.1"
)

// Permission tags used in transaction family names. With these sys admin can configure role level permissions as opposed to account level only. Important Note: as of this version, they must remain empty strings.
const (
	RecipientPermissionTag string = ""
	InitiatorPermissionTag string = ""
)

// Note if payload type is not a key here then "" is returned when we attempt to retrieve the value
var typeToPermissionTag = map[string]string{
	"set_recipient":               RecipientPermissionTag,
	"remove_recipient":            RecipientPermissionTag,
	"set_initiator_rule":          InitiatorPermissionTag,
	"delete_initiator_rule":       InitiatorPermissionTag,
	"add_initiator_to_group":      InitiatorPermissionTag,
	"remove_initiator_from_group": InitiatorPermissionTag,
	"set_initiator_pub_keys":      InitiatorPermissionTag,
	"delete_initiator_pub_keys":   InitiatorPermissionTag,
	"set_account_level_rule":      InitiatorPermissionTag,
	"delete_account_level_rule":   InitiatorPermissionTag,
}

// address calculation related constants
const (
	actorLength = 40
	fieldLength = 20
)
