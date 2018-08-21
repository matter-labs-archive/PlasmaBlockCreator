package types

import (
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// A BigInt represents a signed multi-precision integer.
type BigInt struct {
	Bigint *big.Int
}

// NewBigint allocates and returns a new BigInt set to x.
func NewBigInt(x int64) *BigInt {
	return &BigInt{big.NewInt(x)}
}

// GetBytes returns the absolute value of x as a big-endian byte slice.
func (bi *BigInt) GetBytes() []byte {
	return bi.Bigint.Bytes()
}

// GetBytes returns the absolute value of x as a big-endian byte slice.
func (bi *BigInt) GetLeftPaddedBytes(length int) ([]byte, error) {
	b := bi.Bigint.Bytes()
	if len(b) > length {
		return nil, errors.New("Byte slice is too long to pad")
	}
	return common.LeftPadBytes(b, length), nil
}

// String returns the value of x as a formatted decimal string.
func (bi *BigInt) String() string {
	return bi.Bigint.String()
}

// GetInt64 returns the int64 representation of x. If x cannot be represented in
// an int64, the result is undefined.
func (bi *BigInt) GetInt64() int64 {
	return bi.Bigint.Int64()
}

// SetBytes interprets buf as the bytes of a big-endian unsigned integer and sets
// the big int to that value.
func (bi *BigInt) SetBytes(buf []byte) {
	bi.Bigint.SetBytes(common.CopyBytes(buf))
}

// SetInt64 sets the big int to x.
func (bi *BigInt) SetInt64(x int64) {
	bi.Bigint.SetInt64(x)
}

// Sign returns:
//
//	-1 if x <  0
//	 0 if x == 0
//	+1 if x >  0
//
func (bi *BigInt) Sign() int {
	return bi.Bigint.Sign()
}

// SetString sets the big int to x.
//
// The string prefix determines the actual conversion base. A prefix of "0x" or
// "0X" selects base 16; the "0" prefix selects base 8, and a "0b" or "0B" prefix
// selects base 2. Otherwise the selected base is 10.
func (bi *BigInt) SetString(x string, base int) {
	bi.Bigint.SetString(x, base)
}

// BigInts represents a slice of big ints.
type BigInts struct{ bigints []*big.Int }

// Size returns the number of big ints in the slice.
func (bi *BigInts) Size() int {
	return len(bi.bigints)
}

// Get returns the Bigint at the given index from the slice.
func (bi *BigInts) Get(index int) (Bigint *BigInt, _ error) {
	if index < 0 || index >= len(bi.bigints) {
		return nil, errors.New("index out of bounds")
	}
	return &BigInt{bi.bigints[index]}, nil
}

// Set sets the big int at the given index in the slice.
func (bi *BigInts) Set(index int, Bigint *BigInt) error {
	if index < 0 || index >= len(bi.bigints) {
		return errors.New("index out of bounds")
	}
	bi.bigints[index] = Bigint.Bigint
	return nil
}

// GetString returns the value of x as a formatted string in some number base.
func (bi *BigInt) GetString(base int) string {
	return bi.Bigint.Text(base)
}
