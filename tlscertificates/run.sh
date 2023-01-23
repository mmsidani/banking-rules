#!/bin/bash
go run generate_cert.go --host 127.0.0.1
mv cert.pem key.pem ../lambda
