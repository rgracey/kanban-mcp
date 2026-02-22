package main

import (
	"fmt"

	"github.com/rgracey/kanban-mcp/internal/config"
)

func main() {
	cfg := config.Load()
	fmt.Printf("Config loaded: %+v\n", cfg)
}
