package sharding

import (
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"regexp"
	"strconv"

	fdb "github.com/apple/foundationdb/bindings/go/src/fdb"
	env "github.com/caarlos0/env"
)

type config struct {
	WorkerConcurrency int      `env:"FDB_CONCURRENCY" envDefault:"50000"`
	WorkerConfigs     []string `env:"SHARDS" envDefault:"0-255->0" envSeparator:":"`
}

var FDBPartitionsPool []*FDBWorker = make([]*FDBWorker, 256)
var shardingRegex = regexp.MustCompile(`^([0-9]*)-([0-9]*)->([0-9]*)$`)

func init() {
	cfg := config{}
	err := env.Parse(&cfg)
	if err != nil {
		log.Printf("%+v\n", err)
		panic(1)
	}
	fmt.Printf("%+v\n", cfg)
	fdb.MustAPIVersion(510)
	err = fdb.StartNetwork()
	if err != nil {
		panic(1)
	}
	partitionsInitialized := 0
	fmt.Println("Initiating sharding FDB workers")
	for _, shard := range cfg.WorkerConfigs {
		fmt.Println("Current shard : " + shard)
		match := shardingRegex.FindStringSubmatch(shard)
		if len(match) != 4 {
			panic(1)
		}
		shardStart, err := strconv.ParseUint(match[1], 10, 8)
		if err != nil {
			panic(1)
		}
		shardEnd, err := strconv.ParseUint(match[2], 10, 8)
		if err != nil {
			panic(1)
		}
		shardNumber, err := strconv.ParseUint(match[3], 10, 64)
		if err != nil {
			panic(1)
		}
		if shardNumber > shardEnd {
			panic(1)
		}
		for j := shardStart; j <= shardEnd; j++ {
			fmt.Println("Prefix number " + strconv.FormatUint(j, 10) + " to shard number " + strconv.FormatUint(shardNumber, 10))
			worker := NewFDBWorker(shardNumber, cfg.WorkerConcurrency)
			if FDBPartitionsPool[j] != nil {
				panic(1)
			}
			FDBPartitionsPool[j] = worker
			partitionsInitialized++
		}
	}
	if partitionsInitialized != 256 {
		panic(1)
	}
}

type FDBWorker struct {
	db                 fdb.Database
	Concurrency        int
	concurrencyChannel chan bool
}

func NewFDBWorker(clusterFileIndex uint64, concurrencyLimit int) *FDBWorker {
	absolutePath, err := filepath.Abs("./")
	if err != nil {
		panic(1)
	}
	clusterFileName := absolutePath + "/shards/fdb-" + strconv.FormatUint(clusterFileIndex, 10) + ".cluster"
	c := make(chan bool, concurrencyLimit)
	cluster, err := fdb.CreateCluster(clusterFileName)
	if err != nil {
		panic(1)
	}
	db, err := cluster.OpenDatabase([]byte("DB"))
	if err != nil {
		panic(1)
	}
	new := &FDBWorker{db, concurrencyLimit, c}
	return new
}

func Transact(shardingKey []byte, tx func(tr fdb.Transaction) (interface{}, error)) (interface{}, error) {
	if len(shardingKey) != 1 {
		return nil, errors.New("Too large sharding key")
	}
	shard := FDBPartitionsPool[int(shardingKey[0])]
	return shard.ShardTransact(tx)
}

func ReadTransact(shardingKey []byte, tx func(tr fdb.ReadTransaction) (interface{}, error)) (interface{}, error) {
	if len(shardingKey) != 1 {
		return nil, errors.New("Too large sharding key")
	}
	shard := FDBPartitionsPool[int(shardingKey[0])]
	return shard.ShardReadTransact(tx)
}

func (w *FDBWorker) ShardTransact(tx func(tr fdb.Transaction) (interface{}, error)) (interface{}, error) {
	w.concurrencyChannel <- true
	defer func() { <-w.concurrencyChannel }()
	return w.db.Transact(tx)
}

func (w *FDBWorker) ShardReadTransact(tx func(tr fdb.ReadTransaction) (interface{}, error)) (interface{}, error) {
	w.concurrencyChannel <- true
	defer func() { <-w.concurrencyChannel }()
	return w.db.ReadTransact(tx)
}
