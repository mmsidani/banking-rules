package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/aws/aws-lambda-go/events"
	flags "github.com/jessevdk/go-flags"

	c "../common"

	"encoding/json"
)

func main() {
	var opts c.PayloadFields
	// // // // // // url := c.APIGateway

	parser := flags.NewParser(&opts, flags.Default)
	remaining, err := parser.Parse()
	if err != nil {
		fmt.Println("Error: parsing command line args failed")
	}

	if len(remaining) > 0 {
		fmt.Println("Warning: extraneous options")
	}

	// TODO handle case where keys file doesn't exist
	privateKey, publicKey := c.GetKeysFromFiles(opts.KeysFile)
	p := c.CreateSignedPayload(&opts, &privateKey, &publicKey)

	pEnc, err := json.Marshal(*p)
	if err != nil {
		panic(err)
	}
	// // // // // // req, err := http.NewRequest("POST", url, bytes.NewBuffer(pEnc))
	// // // // // // if err != nil {
	// // // // // // 	panic("http failed")
	// // // // // // }

	// // // // // // req.Header.Set("X-Custom-Header", opts.RequestType)
	// // // // // // req.Header.Set("Content-Type", "application/json")

	// // // // // // client := &http.Client{}
	// // // // // // resp, err := client.Do(req)
	// // // // // // if err != nil {
	// // // // // // 	panic(err)
	// // // // // // }
	// // // // // // defer resp.Body.Close()

	// // // // // // fmt.Println("response Status:", resp.Status)
	// // // // // // fmt.Println("response Headers:", resp.Header)
	// // // // // // body, _ := ioutil.ReadAll(resp.Body)
	// // // // // // fmt.Println("response Body:", string(body))

	var request events.APIGatewayProxyRequest
	request.Body = string(pEnc)
	reqEnc, er := json.Marshal(request)

	if er != nil {
		panic("could not marshal")
	}
	e := ioutil.WriteFile("/home/majed/Documents/code/banking/event.json", reqEnc, os.ModePerm)
	if e != nil {
		panic(e)
	}

}
