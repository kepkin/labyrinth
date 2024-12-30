package labyrinth

type BaseCell struct {
	pos        Position
	typeV      CellType
	Attributes any
}

func (c BaseCell) Type() CellType {
	return c.typeV
}

func (c BaseCell) Pos() Position { return c.pos }

type EarthCell struct {
	pos Position
}

func (c EarthCell) Pos() Position {
	return c.pos
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

func (c WallCell) Pos() Position { return c.pos }

func (c WallCell) Type() CellType {
	return CellType{
		Class: "wall",
		Name:  "wall",
	}
}

type ExitCell struct {
	pos Position
}

func (c ExitCell) Pos() Position { return c.pos }

func (c ExitCell) Type() CellType {
	return CellType{
		Class: "exit",
		Name:  "exit",
	}
}

type RiverCell struct {
	pos     Position
	Dir     MoveDirection
	isMouth bool
}

const RiverCellDirectionAttr = "river_dir"
const RiverCellMouthAttr = "river_mouth"

func (c RiverCell) Pos() Position { return c.pos }

func (c RiverCell) Type() CellType {
	name := "river"
	mouthAttr := ""
	if c.isMouth {
		name = "river mouth"
		mouthAttr = "true"
	}
	return CellType{
		Class:      "river",
		Name:       name,
		Attributes: map[string]string{RiverCellDirectionAttr: c.Dir.Utf8Arrow(), RiverCellMouthAttr: mouthAttr},
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
