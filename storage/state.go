package storage

import (
	"encoding/json"
	"fmt"
	"os"
)

type State struct {
	LatestBlock uint64 `json:"latest_block"`
}

func ReadLatestBlock(filename string) (uint64, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return 0, nil
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return 0, fmt.Errorf("read state file: %w", err)
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return 0, fmt.Errorf("unmarshal state: %w", err)
	}

	return state.LatestBlock, nil
}

func WriteLatestBlock(filename string, block uint64) error {
	state := State{LatestBlock: block}
	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("write state file: %w", err)
	}
	return nil
}