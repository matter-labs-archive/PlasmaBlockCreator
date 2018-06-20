#!/bin/bash
cd ./vendor/github.com/ethereum/go-ethereum/crypto/
rm signature_cgo.go
mv signature_nocgo.go signature.go