package labyrinth

import (
	"fmt"
	"slices"
)

type Command interface {
	Do(w *World, p *Player) []Event
}

type Cell = *CellType

type Item struct {
	ID   HandItem
	Name string
}
type CellType struct {
	Class      string
	Name       string
	Attributes map[string]string
	Items      []*Item

	Custom any
}

func (c *CellType) TakeItem(name string) *Item {
	var result *Item
	c.Items = slices.DeleteFunc(c.Items, func(e *Item) bool {
		if e.Name != name {
			return false
		}

		result = e
		return true
	})

	return result
}

func (c *CellType) PutItem(e *Item) {
	c.Items = append(c.Items, e)
}

func GetXFromLetterMust(l rune) int {
	if l >= 'A' && l <= 'Z' {
		return int(l - 'A')
	}

	if l >= 'a' && l <= 'z' {
		return int(l - 'a')
	}

	if l >= 'а' && l <= 'я' {
		return int(l - 'а')
	}

	if l >= 'А' && l <= 'Я' {
		return int(l - 'А')
	}

	panic("unssupported")
}

type Position struct {
	X int
	Y int
}

func NewPosition(x, y int) Position {
	return Position{X: x, Y: y}
}

func (p Position) String() string {
	return fmt.Sprintf("%v:%v", p.X, p.Y)
}

func (p Position) Next(d MoveDirection) Position {

	switch d {
	case North:
		return Position{
			X: p.X,
			Y: p.Y - 1,
		}
	case East:
		return Position{
			X: p.X + 1,
			Y: p.Y,
		}
	case South:
		return Position{
			X: p.X,
			Y: p.Y + 1,
		}
	case West:
		return Position{
			X: p.X - 1,
			Y: p.Y,
		}
	}

	return p
}

type Size struct {
	Width  int
	Height int
}

type MoveDirection int

func (m MoveDirection) TurnBack() MoveDirection {
	switch m {
	case North:
		return South
	case South:
		return North
	case West:
		return East
	case East:
		return West
	}

	return MoveNil
}

func (m MoveDirection) Utf8Arrow() string {
	switch m {
	case North:
		return "↑"
	case South:
		return "↓"
	case West:
		return "←"
	case East:
		return "→"
	}

	return ""
}

func (m MoveDirection) String() string {
	switch m {
	case North:
		return "north"
	case South:
		return "south"
	case West:
		return "west"
	case East:
		return "east"
	}

	return ""
}

func MoveDirectionFromUtf8Arrow(ch string) MoveDirection {
	switch ch {
	case "←":
		return West
	case "↑":
		return North
	case "→":
		return East
	case "↓":
		return South
	}

	return MoveNil
}

func MoveDirectionFromWord(ch string) (MoveDirection, error) {
	switch ch {
	case "north", "North":
		return North, nil
	case "east", "East":
		return East, nil
	case "south", "South":
		return South, nil
	case "west", "West":
		return West, nil
	}

	return MoveNil, fmt.Errorf("unknown word for move")
}

const (
	MoveNil MoveDirection = iota
	North
	East
	South
	West
)

type EventType int

const (
	MoveEventType = iota
	WinEventType
	ExitEventType
	LearnCellEventType
	RiverDragEventType
	PickObjectEventType
	DropObjectEventType
	LooseObjectEventType
	ErrorEventType
	FoundObjectEventType
	RevealObjectEventType
	TeleportEventType
	GameStartEventType
)

type Event struct {
	Type    EventType
	Subject string
	Value   string
}

func NewEventf2(eventType EventType, subject string, value string) Event {
	return Event{
		Type:    eventType,
		Subject: subject,
		Value:   value,
	}
}

type EventStringer interface {
	ToString(ev Event) string
}

type DefaultEventStringer struct {
}

func (DefaultEventStringer) ToString(ev Event) string {
	switch ev.Type {
	case MoveEventType:
		return fmt.Sprintf("Player %v moved %v", ev.Subject, ev.Value)

	case WinEventType:
		return fmt.Sprintf("Player %v WINS", ev.Subject)

	case ExitEventType:
		return fmt.Sprintf("Player %v found exit", ev.Subject)

	case LearnCellEventType:
		return fmt.Sprintf("Player %v is on %v", ev.Subject, ev.Value)

	case RiverDragEventType:
		return fmt.Sprintf("Player %v dragged downstream", ev.Subject)

	case PickObjectEventType:
		return fmt.Sprintf("Player %v picked up %v", ev.Subject, ev.Value)

	case DropObjectEventType:
		return fmt.Sprintf("Player %v dropped %v", ev.Subject, ev.Value)

	case LooseObjectEventType:
		return fmt.Sprintf("Player %v loosed %v", ev.Subject, ev.Value)

	case ErrorEventType:
		return "Unexpected error"

	case FoundObjectEventType:
		return fmt.Sprintf("Player %v found %v", ev.Subject, ev.Value)

	case RevealObjectEventType:
		if ev.Value == "genuine" {
			return "Player's treasure is genuine"
		}
		return "Player's treasure is fake"

	case TeleportEventType:
		return fmt.Sprintf("Player %v was teleported", ev.Subject)

	case GameStartEventType:
		return fmt.Sprintf("Game started. Player %v is the first to move", ev.Subject)

	}

	return "Unsupported event"
}
