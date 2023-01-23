package tprocessor

import (
	"encoding/json"
	"reflect"
	"syscall"

	c "../common"
	"../core"

	"github.com/hyperledger/sawtooth-sdk-go/processor"
	"github.com/hyperledger/sawtooth-sdk-go/protobuf/processor_pb2"
)

var familyNameReferenceCount = make(map[string]int, 0)
var familyNameToProcessor = make(map[string]*processor.TransactionProcessor, 0)

// Opts is for parsing command line options
type Opts struct {
	FamilyName string `short:"f" long:"family_name" description:"transaction family name, the source account number at this time"`
}

// TpHandler is the handler to register with the validator
type TpHandler struct {
	TpName    string // this is to be used as the transactions family name
	TpVersion string // this the family version
}

// FamilyName returns prefix for transactions handled by this handler
func (r *TpHandler) FamilyName() string {
	return r.TpName
}

// FamilyVersions as the name says...
func (r *TpHandler) FamilyVersions() []string {
	return []string{r.TpVersion}
}

// Namespaces for transactions
func (r *TpHandler) Namespaces() []string {
	return []string{core.Namespace(r.TpName)}
}

// Apply is the main logic
func (r *TpHandler) Apply(tx *processor_pb2.TpProcessRequest, context *processor.Context) error {
	var p c.SignedPayload
	err := json.Unmarshal(tx.Payload, &p)
	if err != nil {
		panic(err)
	}

	t := core.PayloadRegistry[p.Type]
	f := reflect.New(t).Elem().MethodByName("Apply")
	if f.IsValid() {
		v := reflect.ValueOf(p.Payload)
		c := reflect.ValueOf(context)
		ar := f.Call([]reflect.Value{v, c})
		e := ar[0].Interface() // TODO TODO debug replaces this e := ar[0].Interface()

		if e != nil {
			return e.(error)
		}

		return nil
	}
	// we should not conceivably get here since we check for Apply() method in lambda handler
	panic("payload in request does not have Apply() method")
}

// Launch launch transaction processor for this family name
func Launch(familyName string) {
	familyNameReferenceCount[familyName]++
	n := familyNameReferenceCount[familyName]
	if n != 1 {
		// already running: we increase reference count by one and return. TODO later implement logic like familyNameReferenceCount == Threshold => launch new transaction processor?
		return
	}

	// we're here therefore a new processor is needed

	// TODO TODO TODO should we check if a transaction processor for this account is already running?
	processor := processor.NewTransactionProcessor(c.ValidatorEndpoint)
	familyNameToProcessor[familyName] = processor

	processor.ShutdownOnSignal(syscall.SIGINT, syscall.SIGTERM)
	p := &TpHandler{
		TpName:    familyName,
		TpVersion: core.FamilyVersion,
	}
	// TODO TODO TODO TODO TODO for debugging for debugging for debugging
	processor.SetThreadCount(1)
	// TODO TODO TODO TODO TODO for debugging for debugging for debugging
	processor.AddHandler(p)
	err := processor.Start()
	if err != nil {
		panic(err)
	}

}

// ShutDown decrement reference count and shutdown transaction processor for family name if 0
func ShutDown(familyName string) {
	familyNameReferenceCount[familyName]--
	n := familyNameReferenceCount[familyName]
	if n == 0 {
		// reference count down to 0 so shut down
		p := familyNameToProcessor[familyName]
		p.Shutdown()
	}

	return
}
