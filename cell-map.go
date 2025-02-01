package labyrinth

import (
	"fmt"
	"io"
	"iter"
	"log"
	"os"
)

type CellMap struct {
	v [][]Cell
}

func (cm *CellMap) Rows() int {
	return len(cm.v)
}

func (cm *CellMap) Cols() int {
	if len(cm.v) != 0 {
		return len(cm.v[0])
	}

	return 0
}

func (cm *CellMap) Insert(c Cell, p Position) {
	for len(cm.v) <= p.Y {
		cm.v = append(cm.v, nil)
	}

	for x := len(cm.v[p.Y]); x <= p.X; x++ {
		cm.v[p.Y] = append(cm.v[p.Y], &CellType{Class: "wall"})
	}

	cm.v[p.Y][p.X] = c
}

func (cm *CellMap) Get(p Position) Cell {
	if p.Y >= len(cm.v) {
		return &CellType{Class: "wall"}
	}
	if p.X >= len(cm.v[p.Y]) {
		return &CellType{Class: "wall"}
	}

	return cm.v[p.Y][p.X]
}

func (cm *CellMap) All() iter.Seq2[Position, Cell] {
	return func(yield func(Position, Cell) bool) {
		for y, row := range cm.v {
			for x, c := range row {
				if !yield(NewPosition(x, y), c) {
					return
				}
			}
		}
	}
}

// Returns cells in specific inner rectangle
func (cm *CellMap) Rect(ltc Position, rbc Position) iter.Seq2[Position, Cell] {
	return func(yield func(Position, Cell) bool) {
		for y, row := range cm.v {
			if y < ltc.Y || y > rbc.Y {
				continue
			}

			for x, c := range row {
				if x < ltc.X || x > rbc.X {
					continue
				}

				if !yield(NewPosition(x, y), c) {
					return
				}
			}
		}
	}
}

func FPrintCellMap(w io.Writer, cellMap CellMap) {
	lastY := -1
	for p, c := range cellMap.All() {
		if p.Y > lastY { // nextrow
			lastY = p.Y
			if lastY != 0 { // exception for first row
				fmt.Println()
				_, err := w.Write([]byte{'\n'})
				if err != nil {
					log.Print(err.Error())
				}
			}
			_, err := w.Write([]byte{'|'})
			if err != nil {
				log.Print(err.Error())
			}
		}

		if c == nil {
			_, err := w.Write([]byte("nil"))
			if err != nil {
				log.Print(err.Error())
			}
		} else {
			_, err := w.Write([]byte(c.Class))
			if err != nil {
				log.Print(err.Error())
			}
		}

		_, err := w.Write([]byte{'|'})
		if err != nil {
			log.Print(err.Error())
		}
	}
}

func PrintCellMap(cellMap CellMap) {
	FPrintCellMap(os.Stdout, cellMap)
}

type PlayerMap struct {
	LeftCorner  Position
	RightCorner Position
	KnonwnCells map[Position]struct{}
}

func NewPlayerMap(pos Position) PlayerMap {
	return PlayerMap{
		LeftCorner:  pos,
		RightCorner: pos,
		KnonwnCells: map[Position]struct{}{pos: struct{}{}},
	}
}

func (cf *PlayerMap) Learn(pos Position) {
	if cf.KnonwnCells == nil {
		cf.KnonwnCells = map[Position]struct{}{}
	}
	cf.KnonwnCells[pos] = struct{}{}

	if pos.X < cf.LeftCorner.X {
		cf.LeftCorner.X = pos.X
	}
	if pos.Y < cf.LeftCorner.Y {
		cf.LeftCorner.Y = pos.Y
	}

	if pos.X > cf.RightCorner.X {
		cf.RightCorner.X = pos.X
	}
	if pos.Y > cf.RightCorner.Y {
		cf.RightCorner.Y = pos.Y
	}
}
