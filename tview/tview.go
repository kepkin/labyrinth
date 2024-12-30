package tview

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	lab "github.com/kepkin/labyrinth"
)

func NewWorldTable(w *lab.World, players []*lab.Player) WorldTable {
	return WorldTable{
		w:  w,
		pl: players,
	}
}

type WorldTable struct {
	tview.TableContentReadOnly

	w  *lab.World
	pl []*lab.Player
}

func (m *WorldTable) GetCell(row, column int) *tview.TableCell {
	var ret *tview.TableCell
	ret = tview.NewTableCell("")

	worldCell := m.w.Cells.Get(lab.Position{X: column, Y: row})

	worldCellT := worldCell.Type()
	switch worldCellT.Class {
	case "wall":
		ret = tview.NewTableCell(" ")
		ret = ret.SetStyle(ret.Style.Foreground(tcell.ColorGreen))
		ret.SetBackgroundColor(tcell.ColorWhite)
	case "river":
		ch := worldCellT.Attributes[lab.RiverCellDirectionAttr]
		ret = tview.NewTableCell(ch)
		ret.SetBackgroundColor(tcell.ColorCornflowerBlue)

		rcell := worldCell.(lab.RiverCell)
		ret.SetText(rcell.Dir.Utf8Arrow())
	case "wormhole":
		ret = tview.NewTableCell(" ")
		ret.SetBackgroundColor(tcell.ColorDarkGreen)
	}

	for idx, p := range m.pl {
		if p.Pos.X == column && p.Pos.Y == row {
			ret.SetText(fmt.Sprintf("%v", idx))
		}
	}

	return ret
}

func (m *WorldTable) GetRowCount() int {
	return m.w.Dimensions().Height
}

func (m *WorldTable) GetColumnCount() int {
	return m.w.Dimensions().Width
}
