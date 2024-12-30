package labyrinth

type HandItem uint

const (
	Nothing HandItem = iota
	Treasure
	FakeTreasure
)

type Player struct {
	Name string
	Pos  Position

	Hand   HandItem
	Lives  int
	Arrows int

	Attrs map[string]string
}

func (p *Player) SetAttr(attr string, value string) {
	if p.Attrs == nil {
		p.Attrs = make(map[string]string)
	}
	p.Attrs[attr] = value
}

func (p *Player) GetAttr(attr string) string {
	return p.Attrs[attr]
}
