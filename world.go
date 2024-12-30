package labyrinth

type World struct {
	Cells CellMap
	ch    chan Event
}

func (w *World) SetChannel(ch chan Event) {
	w.ch = ch
}

// return X, Y dimensions
func (w *World) Dimensions() Size {
	ret := Size{}
	ret.Height = w.Cells.Rows()
	if ret.Height == 0 {
		return ret
	}
	ret.Width = w.Cells.Cols()

	return ret
}

func (w *World) Emmit(e Event) {
	if w.ch != nil {
		w.ch <- e
	}
}

func NewWorld(wmap [][]string) *World {
	cf := CellWorldBuilder{
		CellFac: DefaultCellFactory,
	}

	for y, row := range wmap {
		for x, cellType := range row {
			cf.MakeCell(cellType, x, y)
		}
	}

	return &World{Cells: cf.CellMap}
}

func NewWorldFromString(wmap string) *World {
	cf := CellWorldBuilder{
		CellFac: DefaultCellFactory,
	}

	var y, x int
	for _, c := range wmap {
		if c == '\n' {
			if x == 0 {
				continue
			}

			y += 1
			x = 0

			continue
		}

		cf.MakeCell(string(c), x, y)
		x += 1
	}

	return &World{Cells: cf.CellMap}
}

func CommandFactory(cmdName string, p *Player, w World) Command {
	switch cmdName {
	case "north":
		return &MoveCommand{
			Direction: North,
		}
	case "east":
		return &MoveCommand{
			Direction: East,
		}
	case "south":
		return &MoveCommand{
			Direction: South,
		}
	case "west":
		return &MoveCommand{
			Direction: West,
		}
	}

	return nil
}
