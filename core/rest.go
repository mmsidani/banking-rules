package core

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	c "../common"
	"github.com/golang/protobuf/proto"
	tpr "github.com/hyperledger/sawtooth-sdk-go/protobuf/transaction_pb2"
)

// Various codes go here
const (
	AddressNotFound int = 404
)

// TODO TODO TODO Note: make it part of the TESTING suite to check that the keys in the payloads constructed in the client are part of VarSet

// StateRespBody struct for state query response body
type StateRespBody struct {
	Data   []map[string]interface{} `json:"Data"`
	Link   string                   `json:"Link"`
	Head   string                   `json:"Head"`
	Paging PagingBody               `json:"Paging"`
	Error  ErrorBody                `json:"Error"`
}

// PagingBody if response broken into pages
type PagingBody struct {
	Start        string `json:"Start"`        // address of first block on page
	Limit        int    `json:"Limit"`        // what we retrieved
	NextPosision string `json:"NextPosition"` // address of block at top of next page
	Next         string `json:"Next"`         // link to next
}

// ErrorBody structure for errors
type ErrorBody struct {
	Code    int    `json:"Code"`
	Title   string `json:"Title"`
	Message string `json:"Message"`
}

// StatusRespBody for status response body
type StatusRespBody struct {
	Data []map[string]interface{} `json:"data"`
	Link string                   `json:"link"`
}

// ParseStateResponse parse http response
func parseStateResponse(resp *http.Response) ([]string, [][]byte) {
	defer resp.Body.Close()

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	body := &StateRespBody{}
	err = json.Unmarshal(buf, body)
	if err != nil {
		panic(err)
	}

	num := len(body.Data)
	addresses := make([]string, num)
	rules := make([][]byte, num)
	for i, r := range body.Data {
		address, ok := r["address"].(string)
		if !ok {
			panic("rest.go address type assertion failed")
		}
		addresses[i] = address
		rulenc, ok := r["data"].(string)
		if !ok {
			panic("rest.go rule type assertion failed")
		}
		rules[i], err = b64.StdEncoding.DecodeString(rulenc)
		if err != nil {
			panic(err)
		}
	}

	return addresses, rules
}

// ParseBatchesResponse parse response from a batch list submission to the rest api
func parseBatchesResponse(resp *http.Response) string {
	defer resp.Body.Close()

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var linkToStatusDict map[string]interface{}
	err = json.Unmarshal(buf, &linkToStatusDict)
	if err != nil {
		// there's an error. in this case linkToStatusDict has one key, "error", and the value is a map with keys "code", "message" and "title". the "message" was not very informative so I didn't bother printing it here
		panic(err)
	}
	// no errors: linkToStatusDict is then a map with one key, "link", with a string value
	linkToStatus := linkToStatusDict["link"].(string)

	return linkToStatus
}

// ParseBatchStatusesResponse returns status of batch submissions
func parseBatchStatusesResponse(resp *http.Response) string {
	defer resp.Body.Close()

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var statusResp StatusRespBody
	err = json.Unmarshal(buf, &statusResp)
	if err != nil {
		panic(err)
	}

	status := statusResp.Data[0]["status"].(string)

	return status
}

// SubmitBatchesReq to rest api
func submitBatchesReq(body []byte) string {
	resp, err := http.Post(c.RestAPIBatches, "application/octet-stream", bytes.NewBuffer(body))
	if err != nil {
		panic(err)
	}
	fmt.Printf("status code after submit batches %d\n", resp.StatusCode)
	linkToStatus := parseBatchesResponse(resp)

	return linkToStatus
}

// SubmitStateReq to rest api
func SubmitStateReq(address string) ([]string, [][]byte) {
	resp, err := http.Get(c.RestAPIState + "?address=" + address)
	if err != nil {
		panic(err)
	}

	if resp.StatusCode == AddressNotFound {
		return nil, nil
	}

	return parseStateResponse(resp)
}

// PollStatus until the submitted batches are no longer pending
func pollStatus(linkToStatus string) string {
	var status string
	for {
		resp, err := http.Get(linkToStatus + "&wait=" + c.RestAPIWait)
		if err != nil {
			panic(err)
		}
		fmt.Printf("status code after poll %d\n", resp.StatusCode)
		status = parseBatchStatusesResponse(resp)
		if status != "PENDING" {
			break
		}
	}

	return status
}

// SubmitTx for submitting transaction through restful API
func SubmitTx(tx *tpr.Transaction) map[string]interface{} {

	// Note: transactor is already being approved/rejected based on Identity Transaction Family data. (logic in validator/server). as part of on-boarding bank (listed in immutable configuration file validator.toml as sole transactor on identity family) will issue transactions to list those at the company that are authorised to transact for specific transaction families, i.e., set/delete rules

	// TODO handle case where keys file doesn't exist
	batchList := createBatchList(tx)

	bEnc, err := proto.Marshal(batchList)
	if err != nil {
		panic(err)
	}

	linkToStatus := submitBatchesReq(bEnc)
	status := pollStatus(linkToStatus)
	if status != "COMMITTED" {
		panic("lambdahandler.go: polling error; status=" + status)
	}

	return map[string]interface{}{
		"link_to_status": []string{linkToStatus},
	}
}
