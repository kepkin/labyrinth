package labyrinth

const (
	CellWall       = "wall"
	CellEarth      = "earth"
	CellRiver      = "river"
	CellRiverMouth = "river mouth"
	CellExit       = "exit"
	CellWormHole   = "wormhole"
)

type SimpleStringCellFactory struct {
	Func func(pos Position) Cell
}

func (sscf SimpleStringCellFactory) Make(key string, pos Position) (Cell, error) {
	return sscf.Func(pos), nil
}

func (rscf SimpleStringCellFactory) Finish(cm CellMap) error {
	return nil
}
