package transaction

const (
	TransactionTypeLength   = 1
	BlockNumberLength       = 4
	TransactionNumberLength = 4
	OutputNumberLength      = 1
	AddressLength           = 20
	ValueLength             = 32
	VLength                 = 1
	RLength                 = 32
	SLength                 = 32

	TransactionTypeSplit = byte(0x01)
	TransactionTypeMerge = byte(0x02)
	TransactionTypeFund  = byte(0x04)
)
