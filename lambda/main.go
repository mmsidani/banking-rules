package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"

	"github.com/hyperledger/sawtooth-sdk-go/protobuf/transaction_pb2"

	c "../common"
	"../core"

	"../tprocessor"
)

func main() {
	// handler for my rest api
	// request resulting in blockchain transaction launches transaction processor and records name in a list. if new transaction on same family_name and family_version comes no new transaction processor is launched. when response sent back (see lambdahandler above) send shutdown signal to transaction processor (don't know how to send signal yet.)

	http.HandleFunc("/", func(w http.ResponseWriter, request *http.Request) {
		var p c.SignedPayload
		body := new(bytes.Buffer)
		body.ReadFrom(request.Body)
		err := json.Unmarshal(body.Bytes(), &p)
		if err != nil {
			panic(err)
		}

		t := p.Type
		s := p.SourceAccount

		fn := core.FamilyName(s, t)

		// launch transaction processor on separate thread because it blocks and polls for messages from validator
		go tprocessor.Launch(fn)

		resp := lambdaHandler(&p)

		b, err := json.Marshal(resp)
		if err != nil {
			panic(err)
		}

		// TODO TODO the request to shut down next is based on the crucial assumption that lambdaHandler() blocks until the query or the transaction completes -- which I believe is very much the case
		tprocessor.ShutDown(fn)

		// TODO needs more work. i'm displaying a json of a map.
		io.WriteString(w, string(b))
	})

	////////log.Printf("About to listen on 8443. Go to https://127.0.0.1:8443/")
	err := http.ListenAndServe(":8443", nil) // TODO upgrade to TLS after figuring out certificates
	// err := http.ListenAndServeTLS(":8443", "cert.pem", "key.pem", nil)
	////////log.Fatal(err)
	if err != nil {
		panic(err)
	}

}

// lambdaHandler is the handler we pass to AWS Lambda. If I've done this right, then new types of transaction can be added without this file being touched. Note: we're relying on the JSON struct tags produced by protoc
func lambdaHandler(p *c.SignedPayload) map[string]string {

	var resp map[string]interface{}
	t := core.PayloadRegistry[p.Type]
	// Look for Handle() method on payload. Handle() is used for querying the state without changing it. One exception, as of today, 9/6/18, is query_auth which, after querying the state, might submit a transaction to the validator to set a "pending bank transaction" if the bank transaction requires multiple signatures
	f := reflect.New(t).Elem().MethodByName("Handle")
	if f.IsValid() {
		v0 := reflect.ValueOf(p.Payload)
		ar := f.Call([]reflect.Value{v0})
		m := ar[0].Interface()
		var ok bool
		resp, ok = m.(map[string]interface{})
		if !ok {
			panic("payload handler failed")
		}
	} else {
		// We're here therefore the payload changes the state and so a transaction has to be submitted to the validator. We construct the transaction out of the payload we received and submit it through the sawtooth rest API
		f = reflect.New(t).Elem().MethodByName("WrapInTx")
		if !f.IsValid() {
			panic("payload is missing a Handle() or WrapInTx() method")
		}

		// After transaction is submitted, the validator passes the request to our transaction processor, which in turn invokes the Apply() method on the payload. Note: g is not needed here. we're only checking for errors
		g := reflect.New(t).Elem().MethodByName("Apply")
		if !g.IsValid() {
			panic("payload must have Apply() method in addition to WrapInTx() method")
		}

		v0 := reflect.ValueOf(p)
		ar := f.Call([]reflect.Value{v0})
		m := ar[0].Interface()
		tx := m.(*transaction_pb2.Transaction)
		resp = core.SubmitTx(tx)
	}

	ret := make(map[string]string, 0)
	for k, v := range resp {
		ret[k] = fmt.Sprintf("%v", v)
	}

	return ret
}
