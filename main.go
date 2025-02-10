package main

import (
	"github.com/realagmag/dictionaryGO/config"
	"github.com/realagmag/dictionaryGO/graph"
)

func main() {
	config.InitDB()
	graph.StartServer()
}
