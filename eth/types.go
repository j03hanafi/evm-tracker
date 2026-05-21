package eth

type TransactionEvent struct {
	TxHash      string
	IsIncoming  bool
	IsNative    bool
	IsInternal  bool
	Amount      string
	TokenAddr   string
	FromAddress string
	ToAddress   string
}
