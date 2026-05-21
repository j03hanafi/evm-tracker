package eth

import (
	"context"
	"math/big"
	
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
)

// TraceCall represents a single call within a transaction trace
type TraceCall struct {
	Type  string      `json:"type"`
	From  string      `json:"from"`
	To    string      `json:"to"`
	Value string      `json:"value"` // Hex encoded
	Calls []TraceCall `json:"calls"`
}

// TraceResult represents the result of tracing a single transaction in a block
type TraceResult struct {
	Result TraceCall `json:"result"`
}

// ExtractInternalTransfers recursively finds all value transfers in a trace
func ExtractInternalTransfers(call TraceCall, txHash string, processor *Processor) []*TransactionEvent {
	var events []*TransactionEvent

	// If this call has a value > 0, it's a native ETH transfer
	if call.Value != "" && call.Value != "0x0" {
		valBytes, err := hexutil.Decode(call.Value)
		if err == nil {
			val := new(big.Int).SetBytes(valBytes)
			from := common.HexToAddress(call.From)
			to := common.HexToAddress(call.To)
			
			if evt := processor.MatchNativeTx(from, to, val, txHash, true); evt != nil {
				events = append(events, evt)
			}
		}
	}

	// Recursively process sub-calls
	for _, subCall := range call.Calls {
		events = append(events, ExtractInternalTransfers(subCall, txHash, processor)...)
	}

	return events
}

// FetchBlockTraces uses debug_traceBlockByNumber to get internal calls
func FetchBlockTraces(ctx context.Context, rpcClient *rpc.Client, blockNumber uint64) ([]TraceResult, error) {
	var traces []TraceResult
	
	blockHex := hexutil.EncodeUint64(blockNumber)
	
	// Use standard callTracer to get internal calls
	err := rpcClient.CallContext(ctx, &traces, "debug_traceBlockByNumber", blockHex, map[string]interface{}{
		"tracer": "callTracer",
	})
	
	if err != nil {
		return nil, err
	}
	
	return traces, nil
}
