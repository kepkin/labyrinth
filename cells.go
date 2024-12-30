package labyrinth

type EarthCell struct {
	pos Position
}

func (c EarthCell) Type() CellType {
	return CellType{
		Class: "earth",
		Name:  "earth",
	}
}

type WallCell struct {
	pos Position
}

func (c WallCell) Type() CellType {
	return CellType{
		Class: "wall",
		Name:  "wall",
	}
}

type ExitCell struct {
	pos Position
}

func (c ExitCell) Type() CellType {
	return CellType{
		Class: "exit",
		Name:  "exit",
	}
}

const BaseCellBuildHookAttr = "build_hook"

type SimpleStringCellFactory struct {
	Func func(pos Position) Cell
}

func (sscf SimpleStringCellFactory) Make(key string, pos Position) (Cell, error) {
	return sscf.Func(pos), nil
}

func (rscf SimpleStringCellFactory) Finish(cm CellMap) {
}
