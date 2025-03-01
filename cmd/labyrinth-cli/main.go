package main

import (
	"image/jpeg"
	"os"

	lab "github.com/kepkin/labyrinth"
	"github.com/kepkin/labyrinth/image"
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

	wimage, err := image.NewCellMapImage(&w.Cells)
	if err != nil {
		panic(err)
	}

	f, err := os.Create("rendered.jpg")
	if err != nil {
		panic(err)
	}
	err = jpeg.Encode(f, wimage, nil)
	if err != nil {
		panic(err)
	}

	gameSession := &lab.Session{
		World:   w,
		Players: pls,
	}

	Run(gameSession)
}
