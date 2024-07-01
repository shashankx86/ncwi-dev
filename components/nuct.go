package main

import (
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func main() {
	app := tview.NewApplication()

	// Title Bar
	title := tview.NewTextView().
		SetText(" nuc ").
		SetTextAlign(tview.AlignCenter).
		SetDynamicColors(true).
		SetTextColor(tcell.ColorGreen).
		SetBackgroundColor(tcell.ColorBlack)

	// Table
	table := tview.NewTable().
		SetBorders(true).
		SetBorder(true).
		SetTitle(" Processes ").
		SetTitleAlign(tview.AlignLeft)
	table.SetCell(0, 0, tview.NewTableCell("PID").SetAlign(tview.AlignCenter).SetSelectable(false).SetTextColor(tcell.ColorYellow))
	table.SetCell(0, 1, tview.NewTableCell("Name").SetAlign(tview.AlignCenter).SetSelectable(false).SetTextColor(tcell.ColorYellow))
	table.SetCell(1, 0, tview.NewTableCell("1").SetAlign(tview.AlignCenter))
	table.SetCell(1, 1, tview.NewTableCell("Alice").SetAlign(tview.AlignCenter))
	table.SetCell(2, 0, tview.NewTableCell("2").SetAlign(tview.AlignCenter))
	table.SetCell(2, 1, tview.NewTableCell("Bob").SetAlign(tview.AlignCenter))

	// Button Bar
	buttonBar := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWrap(false).
		SetText("[green](ctrl+x) Kill  [yellow](ctrl+s) Stop  [red](ctrl+r) Restart").
		SetTextAlign(tview.AlignCenter).
		SetBackgroundColor(tcell.ColorBlack).
		SetBorder(true).
		SetBorderPadding(0, 0, 1, 1)

	// Layout
	layout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(title, 3, 1, false).
		AddItem(table, 0, 1, true).
		AddItem(buttonBar, 3, 1, false)

	buttonBar.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlX, tcell.KeyCtrlS, tcell.KeyCtrlR:
			showHelloWorld(app, layout)
		}
		return event
	})

	if err := app.SetRoot(layout, true).Run(); err != nil {
		panic(err)
	}
}

func showHelloWorld(app *tview.Application, previousRoot tview.Primitive) {
	modal := tview.NewModal().
		SetText("Hello World").
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			app.SetRoot(previousRoot, true).SetFocus(previousRoot)
		})

	app.SetRoot(modal, false).Draw()
	time.AfterFunc(2*time.Second, func() {
		app.QueueUpdateDraw(func() {
			app.SetRoot(previousRoot, true)
		})
	})
}
