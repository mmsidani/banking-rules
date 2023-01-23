package common

// TODO Still have to figure out how to impose authentication on REST API
const (
	ValidatorEndpoint     string = "tcp://localhost:4004"
	RestAPIBatches        string = "http://127.0.0.1:8008/batches"
	RestAPIState          string = "http://127.0.0.1:8008/state"
	RestAPIBatchStatuses  string = "http://127.0.0.1:8008/batch_statuses"
	RestAPIWait           string = "300"
	APIGateway            string = "http://127.0.0.1:3000/"
	AuthUser              string = ""
	AuthPassword          string = ""
	BatchSignerKeysFile   string = "/home/majed/.sawtooth/keys/bank"
	BatchSignerPubKeyFile string = "/home/majed/.sawtooth/keys/bank.pub"
)

// Every ID authorized to use the account is part of this group. This is useful for setting account level rules for sign-offs, etc.,
const (
	DefaultGroupName string = "Everyone" // this is used to set account level rules. Everyone belongs to this group.
)

// RuleVariablesSet defines the variables we expect in an expression. TODO how exhaustive can we be? TODO how are our functions treated?
var RuleVariablesSet = map[string]bool{ // Keep it in alphabetical order for easy reading
	"Action":        true,
	"Amount":        true,
	"Balance":       true,
	"Destaccount":   true,
	"Initiator":     true,
	"Recipient":     true,
	"Rule":          true,
	"Ruletype":      true, // this is not used currently, right?
	"Sourceaccount": true,
}
