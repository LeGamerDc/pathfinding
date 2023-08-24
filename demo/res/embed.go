package res

import (
	_ "embed"
)

var (
	//go:embed agent.png
	AgentPng []byte

	//go:embed blocks.png
	BlockPng []byte
)
