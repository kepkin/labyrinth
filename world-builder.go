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
	//TODO: `pos` parameter is needed only for wormohle. Maybe we can refactor wormhole factory to use pos in Finish only?
	Make(key string, pos Position) (Cell, error)
	Finish(cm CellMap) error
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

func (cf *PrefixChainCellFactory) Finish(cm CellMap) error {
	for _, factory := range cf.facMap {
		err := factory.Finish(cm)
		if err != nil {
			return err
		}
	}

	return nil
}

var DefaultCellFactory *PrefixChainCellFactory

func init() {
	DefaultCellFactory = &PrefixChainCellFactory{}

	_ = DefaultCellFactory.Register(
		[]string{"", " "},
		SimpleStringCellFactory{func(pos Position) Cell { return &CellType{Class: CellEarth} }},
	)
	_ = DefaultCellFactory.Register(
		[]string{CellWall, "w"},
		SimpleStringCellFactory{func(pos Position) Cell { return &CellType{Class: CellWall} }},
	)
	_ = DefaultCellFactory.Register(
		[]string{CellExit, "e"},
		SimpleStringCellFactory{func(pos Position) Cell { return &CellType{Class: CellExit} }},
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
