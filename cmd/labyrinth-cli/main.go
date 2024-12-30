package main

import (
	"os"

	lab "github.com/kepkin/labyrinth"
	md "github.com/kepkin/labyrinth/markdown"
)

func main() {
	if len(os.Args) != 2 {
		panic("use: ./labyrinth map.md")
	}

	b, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err.Error())
	}
	bb := md.WorldBuilder{
		Cf: lab.CellWorldBuilder{
			CellFac: lab.DefaultCellFactory,
		},
	}
	w, pls, err := bb.Build(string(b))
	if err != nil {
		panic(err.Error())
	}

	gameSession := &lab.Session{
		World:   w,
		Players: pls,
	}

	Run(gameSession)
}
