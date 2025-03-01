package main

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	lab "github.com/kepkin/labyrinth"
	labtv "github.com/kepkin/labyrinth/tview"
)

func Run(gameSession *lab.Session) {
	w := gameSession.World
	players := gameSession.Players

	worldChannel := make(chan lab.Event, 10)
	w.SetChannel(worldChannel)
	// background := tview.NewTextView().
	// 	SetTextColor(tcell.ColorBlue).
	// 	SetText(strings.Repeat("background ", 1000))

	tb := tview.NewTable()
	tb.SetBackgroundColor(tcell.ColorDefault)

	mtc := labtv.NewWorldTable(w, players)
	tb.SetContent(&mtc)

	app := tview.NewApplication()

	dropdown := tview.NewDropDown().SetLabel("Choose next move: ").
		SetOptions([]string{"North", "East", "South", "West"}, nil)

	dropdown.SetBackgroundColor(tcell.ColorDefault)

	logView := tview.NewTextView()
	logView.SetDynamicColors(true).SetBackgroundColor(tcell.ColorDefault)
	logView.SetRegions(true)

	posView := tview.NewTextView()
	posView.SetDynamicColors(true).SetBackgroundColor(tcell.ColorDefault)

	var setOptions func(actions []string)

	dropdownSelFunc := func(text string, index int) {
		gameSession.Do(text)
		actions := gameSession.GetCurrentPlayerPossibleActions()
		setOptions(actions)
	}

	setOptions = func(actions []string) {
		dropdown.SetOptions(actions, dropdownSelFunc)
	}

	dropdown.SetSelectedFunc(dropdownSelFunc)

	vFlex := tview.NewFlex().SetDirection(tview.FlexRow)
	vFlex.AddItem(tb, 10, 0, false)
	vFlex.AddItem(dropdown, 3, 0, true)
	vFlex.AddItem(dropdown, 3, 0, false)

	hFlex := tview.NewFlex()
	hFlex.AddItem(vFlex, 25, 0, true)
	hFlex.AddItem(posView, 0, 1, false)
	hFlex.AddItem(logView, 0, 1, false)

	worldEventHandler := func(event lab.Event) {
		app.QueueUpdateDraw(func() {
			fmt.Fprint(logView, LabEventToString(event)+"\n")
			logView.ScrollToEnd()

			posView.Clear()
			for _, p := range players {
				fmt.Fprintf(posView, "player %v - %s\n", p.Name, p.Pos)
			}

			if event.Type == lab.WinEventType {
				app.Stop()
			}
		})
	}

	go func() {
		worldEventHandler(lab.NewEventf2(lab.GameStartEventType, "", ""))
		for event := range worldChannel {
			worldEventHandler(event)

		}
	}()

	if err := app.SetRoot(hFlex, true).Run(); err != nil {
		panic(err)
	}
}

func LabEventToString(ev lab.Event) string {
	switch ev.Type {
	case lab.MoveEventType:
		return fmt.Sprintf("Player %v moved %v", ev.Subject, ev.Value)

	case lab.WinEventType:
		return fmt.Sprintf("Player %v WINS", ev.Subject)

	case lab.ExitEventType:
		return fmt.Sprintf("Player %v found exit", ev.Subject)

	case lab.LearnCellEventType:
		return fmt.Sprintf("Player %v is on %v", ev.Subject, ev.Value)

	case lab.RiverDragEventType:
		return fmt.Sprintf("Player %v dragged downstream", ev.Subject)

	case lab.PickObjectEventType:
		return fmt.Sprintf("Player %v picked up %v", ev.Subject, ev.Value)

	case lab.DropObjectEventType:
		return fmt.Sprintf("Player %v dropped %v", ev.Subject, ev.Value)

	case lab.LooseObjectEventType:
		return fmt.Sprintf("Player %v loosed %v", ev.Subject, ev.Value)

	case lab.ErrorEventType:
		return fmt.Sprint("Unexpected error")

	case lab.FoundObjectEventType:
		return fmt.Sprintf("Player %v found %v", ev.Subject, ev.Value)

	case lab.RevealObjectEventType:
		if ev.Value == "genuine" {
			return fmt.Sprintf("Player's treasure is genuine")
		}
		return fmt.Sprintf("Player's treasure is fake")

	case lab.TeleportEventType:
		return fmt.Sprintf("Player %v was teleported", ev.Subject)

	case lab.GameStartEventType:
		return fmt.Sprintf("Game started. Player %v is the first to move", ev.Subject)

	}

	return "Unsupported event"
}
