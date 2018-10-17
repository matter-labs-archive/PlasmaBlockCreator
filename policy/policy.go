package policy

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"math/big"
	"os"

	"github.com/caarlos0/env"
	common "github.com/ethereum/go-ethereum/common"
	"github.com/matterinc/PlasmaCommons/transaction"
)

type PolicyConfig struct {
	FeeBeneficiary string `env:"FEE_ADDRESS" envDefault:"0x6394b37cf80a7358b38068f0ca4760ad49983a1b"`
	FeeAmount      string `env:"FEE_PER_BRANCH" envDefault:"0"`
	UtxoSize       string `env:"MIN_UTXO_SIZE" envDefault:"1000000000000"` // 1 million part of ETH
}

type Policy struct {
	For             common.Address
	AmountPerBranch *big.Int
	UtxoSize        *big.Int
}

var zero = big.NewInt(0)
var defaultUtxoSize, _ = big.NewInt(0).SetString("100000000000", 10)

var policy = Policy{
	common.HexToAddress("0x6394b37cf80a7358b38068f0ca4760ad49983a1b"),
	big.NewInt(0),
	defaultUtxoSize,
}

func init() {
	cfg := PolicyConfig{}
	err := env.Parse(&cfg)
	if err != nil {
		log.Printf("%+v\n", err)
		os.Exit(1)
	}
	fmt.Printf("%+v\n", cfg)
	feeBeneficiary := common.HexToAddress(cfg.FeeBeneficiary)
	feeAmount, success := big.NewInt(0).SetString(cfg.FeeAmount, 10)
	if !success {
		log.Println("Can not parse fee amount")
		os.Exit(1)
	}
	utxoSize, success := big.NewInt(0).SetString(cfg.UtxoSize, 10)
	if !success {
		log.Println("Can not parse fee amount")
		os.Exit(1)
	}
	policy.For = feeBeneficiary
	policy.AmountPerBranch = feeAmount
	policy.UtxoSize = utxoSize
}

func CheckForPolicy(tx *transaction.SignedTransaction) error {
	err := checkUTXOsizes(tx)
	if err != nil {
		return err
	}
	err = checkFee(tx)
	return err
}

func checkUTXOsizes(tx *transaction.SignedTransaction) error {
	if tx.UnsignedTransaction.TransactionType[0] == transaction.TransactionTypeFund {
		return nil
	}
	for _, output := range tx.UnsignedTransaction.Outputs {
		utxoSizeBytes := output.Value
		utxoSize := big.NewInt(0).SetBytes(utxoSizeBytes[:])
		if utxoSize.Cmp(policy.UtxoSize) == -1 {
			return errors.New("Output size is too small")
		}
	}
	return nil
}

func checkFee(tx *transaction.SignedTransaction) error {
	if tx.UnsignedTransaction.TransactionType[0] == transaction.TransactionTypeFund {
		return nil
	} else if policy.AmountPerBranch.Cmp(zero) == 0 {
		return nil
	} else if tx.UnsignedTransaction.TransactionType[0] == transaction.TransactionTypeMerge {
		numInputs := len(tx.UnsignedTransaction.Inputs)
		numOutputs := len(tx.UnsignedTransaction.Outputs)
		if numInputs > 1 && numOutputs == 1 {
			return nil
		}
		return errors.New("Merging transaction does not reduce a branching factor")
	} else if tx.UnsignedTransaction.TransactionType[0] == transaction.TransactionTypeSplit {
		if policy.AmountPerBranch.Cmp(zero) == 0 {
			return nil
		}
		numInputs := len(tx.UnsignedTransaction.Inputs)
		numOutputs := len(tx.UnsignedTransaction.Outputs)
		if numOutputs < 2 {
			return errors.New("Split transaction with less than 2 outputs and non-zero fee policy")
		}
		lastOutput := tx.UnsignedTransaction.Outputs[numOutputs-1]
		branchingFactor := numInputs - (numOutputs - 1)
		if branchingFactor <= 0 {
			return errors.New("Split transaction with negativa branching, should be Merge instead")
		}
		if bytes.Compare(lastOutput.To[:], policy.For[:]) != 0 {
			return errors.New("Invalid fee recipient")
		}
		expectedFee := big.NewInt(int64(branchingFactor))
		expectedFee = expectedFee.Mul(expectedFee, policy.AmountPerBranch)
		if expectedFee.Cmp(lastOutput.GetValue().Bigint) < 0 {
			return errors.New("Fee is too small")
		}
		return nil
	}
	return errors.New("Invalid transaction type")
}
