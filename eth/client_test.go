package eth

import (
	"encoding/json"
	"testing"
)

func TestTraceCallResultUnmarshal(t *testing.T) {
	// Simulated output from debug_traceBlockByNumber with callTracer
	jsonData := `[
		{
			"result": {
				"type": "CALL",
				"from": "0x1111111111111111111111111111111111111111",
				"to": "0x2222222222222222222222222222222222222222",
				"value": "0x3e8",
				"calls": [
					{
						"type": "CALL",
						"from": "0x2222222222222222222222222222222222222222",
						"to": "0x3333333333333333333333333333333333333333",
						"value": "0x1f4"
					}
				]
			}
		}
	]`

	var traces []TraceResult
	if err := json.Unmarshal([]byte(jsonData), &traces); err != nil {
		t.Fatalf("Failed to unmarshal trace: %v", err)
	}

	if len(traces) != 1 {
		t.Fatalf("Expected 1 trace, got %d", len(traces))
	}
	
	if traces[0].Result.From != "0x1111111111111111111111111111111111111111" {
		t.Errorf("Expected from address, got %s", traces[0].Result.From)
	}
	
	if len(traces[0].Result.Calls) != 1 {
		t.Fatalf("Expected 1 subcall, got %d", len(traces[0].Result.Calls))
	}
	
	if traces[0].Result.Calls[0].Value != "0x1f4" {
		t.Errorf("Expected subcall value 0x1f4, got %s", traces[0].Result.Calls[0].Value)
	}
}
