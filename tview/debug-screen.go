package tview

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	lab "github.com/kepkin/labyrinth"
)

func RunDebug(w *lab.World, players []*lab.Player) {

	worldChannel := make(chan lab.Event, 10)
	w.SetChannel(worldChannel)

	tb := tview.NewTable()
	tb.SetBackgroundColor(tcell.ColorDefault)

	mtc := NewWorldTable(w, players)
	tb.SetContent(&mtc)

	app := tview.NewApplication()

	logView := tview.NewTextView()
	logView.SetDynamicColors(true).SetBackgroundColor(tcell.ColorDefault)
	logView.SetRegions(true)

	posView := tview.NewTextView()
	posView.SetDynamicColors(true).SetBackgroundColor(tcell.ColorDefault)

	vFlex := tview.NewFlex().SetDirection(tview.FlexRow)
	vFlex.AddItem(tb, 10, 0, false)
	vFlex.AddItem(posView, 10, 0, false)
	vFlex.AddItem(logView, 0, 1, false)

	eventStringer := lab.DefaultEventStringer{}

	go func() {
		for event := range worldChannel {
			app.QueueUpdateDraw(func() {
				fmt.Fprint(logView, eventStringer.ToString(event)+"\n")
				logView.ScrollToEnd()

				posView.Clear()
				for _, pl := range players {
					fmt.Fprintf(posView, "player pos %s", pl.Pos)
				}

			})
		}
	}()

	if err := app.SetRoot(vFlex, true).Run(); err != nil {
		panic(err)
	}
}
