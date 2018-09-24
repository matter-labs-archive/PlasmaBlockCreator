package configs

import (
	"errors"
	"fmt"
	"log"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/caarlos0/env"
)

type HTTPConfig struct {
	Port                int `env:"PORT" envDefault:"3001"`
	HTTPConcurrency     int `env:"HTTP_CONCURRENCY" envDefault:"50000"`
	MaxConnectionsPerIP int `env:"HTTP_MAXCONNECTIONS" envDefault:"50000"`
	MaxBodySize         int `env:"HTTP_MAXBODYSIZE" envDefault:"5000"`
}

type RedisConfig struct {
	RedisHost     string `env:"REDIS_HOST" envDefault:"127.0.0.1"`
	RedisPort     int    `env:"REDIS_PORT" envDefault:"6379"`
	RedisPassword string `env:"REDIS_PASSWORD" envDefault:""`
}

type ConcurrencyConfig struct {
	DatabaseConcurrency  int `env:"FDB_CONCURRENCY" envDefault:"100000"`
	ECRecoverConcurrency int `env:"EC_CONCURRENCY" envDefault:"-1"`
	MaxProc              int `env:"GOMAXPROCS" envDefault:"-1"`
}

type FDBConfig struct {
	FdbRewriteClusterFile bool   `env:"FDB_REWRITE" envDefault:"false"`
	FdbClusterFilePath    string `env:"FDB_CLUSTER_FILE_PATH" envDefault:""`
}

type SignatureConfig struct {
	FundingTXSigningKey string `env:"FUNDINGTX_ETH_KEY" envDefault:"0xc87509a1c067bbde78beb793e6fa76530b6382a4c0241e5e4a9ec0a0f44dc0d3"`
	BlockSigningKey     string `env:"BLOCK_ETH_KEY" envDefault:"0xc87509a1c067bbde78beb793e6fa76530b6382a4c0241e5e4a9ec0a0f44dc0d3"`
}

func ParseConfigs() (*HTTPConfig, *RedisConfig, *ConcurrencyConfig, *FDBConfig, *SignatureConfig, error) {
	httpConfig := HTTPConfig{}
	err := env.Parse(&httpConfig)
	if err != nil {
		log.Printf("%+v\n", err)
		return nil, nil, nil, nil, nil, err
	}
	fmt.Printf("%+v\n", httpConfig)

	redisConfig := RedisConfig{}
	err = env.Parse(&redisConfig)
	if err != nil {
		log.Printf("%+v\n", err)
		return nil, nil, nil, nil, nil, err
	}
	fmt.Printf("%+v\n", redisConfig)

	concurrencyConfig := ConcurrencyConfig{}
	err = env.Parse(&concurrencyConfig)
	if err != nil {
		log.Printf("%+v\n", err)
		return nil, nil, nil, nil, nil, err
	}
	fmt.Printf("%+v\n", concurrencyConfig)

	databaseConfig := FDBConfig{}
	err = env.Parse(&databaseConfig)
	if err != nil {
		log.Printf("%+v\n", err)
		return nil, nil, nil, nil, nil, err
	}
	fmt.Printf("%+v\n", databaseConfig)

	signatureConfig := SignatureConfig{}
	err = env.Parse(&signatureConfig)
	if err != nil {
		log.Printf("%+v\n", err)
		return nil, nil, nil, nil, nil, err
	}
	fmt.Printf("%+v\n", signatureConfig)

	return &httpConfig, &redisConfig, &concurrencyConfig, &databaseConfig, &signatureConfig, nil

}

func InitDB(config *FDBConfig) (*fdb.Database, error) {
	err := fdb.StartNetwork()
	if err != nil {
		return nil, err
	}
	if config.FdbRewriteClusterFile == false {
		db := fdb.MustOpenDefault()
		return &db, nil
	}
	if config.FdbClusterFilePath == "" {
		return nil, errors.New("Empty content for cluster file rewriting")
	}

	clusterFileName := config.FdbClusterFilePath
	cluster, err := fdb.CreateCluster(clusterFileName)
	if err != nil {
		return nil, err
	}
	db, err := cluster.OpenDatabase([]byte("DB"))
	if err != nil {
		return nil, err
	}
	return &db, nil
}
