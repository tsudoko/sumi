package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/andlabs/ui"
)

type CboxFunc func(*ui.Combobox)

type ValueCbox struct {
	values         []rune
	cbox           *ui.Combobox
	CallOnSelected func(*ui.Combobox)
}

func (c *ValueCbox) Selected() rune {
	return c.values[c.cbox.Selected()]
}

func NewValueCbox(matches []rune) *ValueCbox {
	vcb := &ValueCbox{matches, ui.NewCombobox(), nil}
	for _, m := range matches {
		vcb.cbox.Append(string(m))
	}
	vcb.cbox.SetSelected(0)
	return vcb
}

func generateBoxes(matches [][]rune) []*ValueCbox {
	boxes := make([]*ValueCbox, 0, 3)
	for _, b := range matches {
		boxes = append(boxes, NewValueCbox(b))
	}

	return boxes
}

func cleanup(path string) {
	os.RemoveAll(path)
}

func cbPrintEntry(e *ui.Entry) func(*ui.Button) {
	return func(*ui.Button) {
		fmt.Println(e.Text())
	}
}

func cbModifyEntry(e *ui.Entry, i int, v *ValueCbox) func(*ui.Combobox) {
	return func(*ui.Combobox) {
		old := e.Text()
		runes := []rune(old)

		for j := len(runes); i >= len(runes); j++ {
			runes = append(runes, '　')
		}

		runes[i] = v.Selected()
		new_ := string(runes)
		e.SetText(new_)
	}
}

func cbSelectArea(w *ui.Window, g *ui.Group, entry *ui.Entry, tempDir string) func(*ui.Button) {
	return func(button *ui.Button) {
		var matches [][]rune

		button.Disable()
		imgPath, err := TakeScreenshot(tempDir+string(os.PathSeparator)+"sumi", "")

		if err != nil {
			ui.MsgBoxError(w, strError, err.Error())
			return
		}

		label := ui.NewLabel(strDetecting_)

		g.SetChild(label)

		go func() {
			matches, err = detectCharacters(imgPath)
			ui.QueueMain(func() {
				label.SetText("")
				button.Enable()

				if err != nil {
					ui.MsgBoxError(w, strError, err.Error())
					return
				}

				box := ui.NewHorizontalBox()
				boxes := generateBoxes(matches)

				entry.SetText("")
				for i, e := range boxes {
					box.Append(e.cbox, false)
					e.cbox.OnSelected(cbModifyEntry(entry, i, e))
					e.CallOnSelected = cbModifyEntry(entry, i, e)
					e.CallOnSelected(e.cbox)
				}

				g.SetChild(box)
			})
		}()
	}
}

func MainWindow() {
	selectButton := ui.NewButton(strSelect)
	printButton := ui.NewButton(strPrint)
	resultEntry := ui.NewEntry()

	matchesGroup := ui.NewGroup("")

	mainbox := ui.NewHorizontalBox()
	otherbox := ui.NewVerticalBox()
	toolbar := ui.NewVerticalBox()

	w := ui.NewWindow("すみ", 0, 0, false)

	tempDir, err := ioutil.TempDir("", "sumi")
	if err != nil {
		ui.MsgBoxError(w, strError, err.Error())
	}

	toolbar.Append(selectButton, false)
	toolbar.Append(printButton, false)
	otherbox.Append(resultEntry, false)
	otherbox.Append(matchesGroup, false)
	// do we really need dynamic box generation? most words are less than 3 chars long
	selectButton.OnClicked(cbSelectArea(w, matchesGroup, resultEntry, tempDir))
	printButton.OnClicked(cbPrintEntry(resultEntry))
	resultEntry.SetReadOnly(true)

	matchesGroup.SetMargined(false)

	mainbox.Append(toolbar, false)
	mainbox.Append(otherbox, false)

	w.SetChild(mainbox)
	w.OnClosing(func(*ui.Window) bool {
		ui.Quit()
		cleanup(tempDir)
		return true
	})

	w.Show()
}

func main() {
	err := ui.Main(MainWindow)

	if err != nil {
		panic(err)
	}
}
