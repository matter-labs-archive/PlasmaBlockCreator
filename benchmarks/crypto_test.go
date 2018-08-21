package benchmarks

import (
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"sync"
	"testing"

	secp256k1 "github.com/shamatar/go-plasma/crypto/secp256k1"
	btcec "github.com/btcsuite/btcd/btcec"
	"github.com/caarlos0/env"
)

func BenchmarkSecp256k1(b *testing.B) {
	hash := make([]byte, 32)
	privateKey := make([]byte, 32)
	rand.Read(hash)
	rand.Read(privateKey)
	signature, _ := secp256k1.Sign(hash, privateKey)
	var wg sync.WaitGroup
	for i := 0; i < b.N; i++ {
		for j := 0; j < 1000000; j++ {
			wg.Add(1)
			go func() {
				_, err := secp256k1.RecoverPubkey(hash, signature)
				if err != nil {
					panic("1")
				}
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

type config struct {
	MaxProc int `env:"GOMAXPROCS" envDefault:"-1"`
}

func BenchmarkGo256k1(b *testing.B) {
	cfg := config{}
	err := env.Parse(&cfg)
	if err != nil {
		log.Printf("%+v\n", err)
		panic(1)
	}
	fmt.Printf("%+v\n", cfg)
	MaxProc := cfg.MaxProc
	if MaxProc == -1 {
		MaxProc = runtime.NumCPU()
	}
	concChannel := make(chan bool, MaxProc)
	hash := make([]byte, 32)
	privateKey := make([]byte, 32)
	rand.Read(hash)
	rand.Read(privateKey)
	curve := btcec.S256()
	privKey, _ := btcec.PrivKeyFromBytes(curve, privateKey)
	signature, _ := btcec.SignCompact(curve, privKey, hash, false)
	var wg sync.WaitGroup
	for i := 0; i < b.N; i++ {
		for j := 0; j < 1000000; j++ {
			wg.Add(1)
			concChannel <- true
			go func() {
				_, _, err := btcec.RecoverCompact(curve, signature, hash)
				if err != nil {
					panic("1")
				}
				<-concChannel
				wg.Done()
			}()
		}
		wg.Wait()
	}
}
