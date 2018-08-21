package secp256k1

import (
	"fmt"
	"log"
	"runtime"
	"strconv"

	env "github.com/caarlos0/env"
)

var WorkerPool Secp256k1WorkerPool

type config struct {
	MaxProc int `env:"GOMAXPROCS" envDefault:"-1"`
}

type Secp256k1WorkerPool struct {
	Concurrency    int
	contextChannel chan *Secp256k1BoundProcessor
}

func init() {
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
	fmt.Println("Initiated SECP256K1 workers: " + strconv.Itoa(MaxProc))
	c := make(chan *Secp256k1BoundProcessor, MaxProc)
	for i := 0; i < MaxProc; i++ {
		proc := NewSecp256k1BoundProcessor()
		c <- proc
	}
	newPool := Secp256k1WorkerPool{Concurrency: MaxProc, contextChannel: c}
	WorkerPool = newPool
}

// Sign creates a recoverable ECDSA signature.
// The produced signature is in the 65-byte [R || S || V] format where V is 0 or 1.
//
// The caller is responsible for ensuring that msg cannot be chosen
// directly by an attacker. It is usually preferable to use a cryptographic
// hash function on any input before handing it to this function.
func Sign(msg []byte, seckey []byte) ([]byte, error) {
	boundContext := <-WorkerPool.contextChannel
	defer func() { WorkerPool.contextChannel <- boundContext }()
	return boundContext.Sign(msg, seckey)
}

// RecoverPubkey returns the the public key of the signer.
// msg must be the 32-byte hash of the message to be signed.
// sig must be a 65-byte compact ECDSA signature containing the
// recovery id as the last element.
func RecoverPubkey(msg []byte, sig []byte) ([]byte, error) {
	boundContext := <-WorkerPool.contextChannel
	defer func() { WorkerPool.contextChannel <- boundContext }()
	return boundContext.RecoverPubkey(msg, sig)
}

// VerifySignature checks that the given pubkey created signature over message.
// The signature should be in [R || S] format.
func VerifySignature(pubkey, msg, signature []byte) bool {
	boundContext := <-WorkerPool.contextChannel
	defer func() { WorkerPool.contextChannel <- boundContext }()
	return boundContext.VerifySignature(pubkey, msg, signature)
}
