package labyrinth

import (
	"fmt"
	"strings"
)

type CellWorldBuilder struct {
	CellMap CellMap
	CellFac StringCellFactory
}

func (cf *CellWorldBuilder) MakeCell(cellType string, x int, y int) {
	c, err := cf.CellFac.Make(strings.TrimSpace(cellType), NewPosition(x, y))
	if err != nil {
		panic(err)
	}
	cf.CellMap.Insert(c, NewPosition(x, y))
}

func (cf *CellWorldBuilder) BuildCellMap() (CellMap, error) {
	cf.CellFac.Finish(cf.CellMap)
	return cf.CellMap, nil
}

type StringCellFactory interface {
	Make(key string, pos Position) (Cell, error)
	Finish(cm CellMap)
}

type PrefixChainCellFactory struct {
	facMap map[string]StringCellFactory
}

func (cf *PrefixChainCellFactory) Register(keys []string, factory StringCellFactory) error {
	if cf.facMap == nil {
		cf.facMap = map[string]StringCellFactory{}
	}

	for _, key := range keys {
		if _, ok := cf.facMap[key]; ok {
			return fmt.Errorf("conflict key")
		}

		cf.facMap[key] = factory
	}

	return nil
}

func (cf *PrefixChainCellFactory) Make(cellCode string, pos Position) (Cell, error) {
	key, _, _ := strings.Cut(cellCode, ":")
	factory, ok := cf.facMap[key]

	if !ok {
		return nil, fmt.Errorf("no factory for this cell %v", cellCode)
	}

	return factory.Make(cellCode, pos)
}

func (cf *PrefixChainCellFactory) Finish(cm CellMap) {
	for _, factory := range cf.facMap {
		factory.Finish(cm)
	}
}

var DefaultCellFactory *PrefixChainCellFactory

func init() {
	DefaultCellFactory = &PrefixChainCellFactory{}

	_ = DefaultCellFactory.Register(
		[]string{"", " "},
		SimpleStringCellFactory{func(pos Position) Cell { return EarthCell{pos} }},
	)
	_ = DefaultCellFactory.Register(
		[]string{"wall", "w"},
		SimpleStringCellFactory{func(pos Position) Cell { return WallCell{pos} }},
	)
	_ = DefaultCellFactory.Register(
		[]string{"exit", "e"},
		SimpleStringCellFactory{func(pos Position) Cell { return ExitCell{pos} }},
	)
	_ = DefaultCellFactory.Register(
		RiverStringFactoryKeys,
		&RiverStringCellFactory{},
	)
	_ = DefaultCellFactory.Register(
		[]string{"W"},
		&WormholeStringCellFactory{},
	)
}
