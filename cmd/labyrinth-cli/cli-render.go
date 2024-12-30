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

	dropdown.SetSelectedFunc(func(text string, index int) {
		gameSession.Do(text)
	})

	vFlex := tview.NewFlex().SetDirection(tview.FlexRow)
	vFlex.AddItem(tb, 10, 0, false)
	vFlex.AddItem(dropdown, 3, 0, true)
	vFlex.AddItem(dropdown, 3, 0, false)

	hFlex := tview.NewFlex()
	hFlex.AddItem(vFlex, 25, 0, true)
	hFlex.AddItem(posView, 0, 1, false)
	hFlex.AddItem(logView, 0, 1, false)

	go func() {
		for logValue := range worldChannel {
			app.QueueUpdateDraw(func() {
				fmt.Fprint(logView, logValue+"\n")
				logView.ScrollToEnd()

				posView.Clear()
				for _, p := range players {
					fmt.Fprintf(posView, "player %v - %s\n", p.Name, p.Pos)
				}

			})
		}
	}()

	if err := app.SetRoot(hFlex, true).Run(); err != nil {
		panic(err)
	}
}
