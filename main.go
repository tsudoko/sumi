package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/GeertJohan/go.tesseract"

	"github.com/mattn/go-gtk/gdk"
	"github.com/mattn/go-gtk/glib"
	"github.com/mattn/go-gtk/gtk"
)

func generateBoxes(matches [][]rune) []*gtk.ComboBoxText {
	boxes := make([]*gtk.ComboBoxText, 0, 3)
	for _, b := range matches {
		cb := gtk.NewComboBoxText()
		for _, m := range b {
			cb.AppendText(string(m))
		}
		cb.SetActive(0)
		boxes = append(boxes, cb)
	}
	return boxes
}

func cleanup(path string) {
	os.RemoveAll(path)
}

func cbModifyEntry(e *gtk.Entry, i int, cbt *gtk.ComboBoxText) func() {
	return func() {
		old := e.GetText()
		runes := []rune(old)

		for j := len(runes); i >= len(runes); j++ {
			runes = append(runes, '　')
		}

		if cbt.GetActiveText() != "" {
			runes[i] = []rune(cbt.GetActiveText())[0]
		}
		new_ := string(runes)
		e.SetText(new_)
	}
}

func cbSelectArea(w *gtk.Window, t *tesseract.Tess, butt *gtk.Button, box *gtk.Box, entry *gtk.Entry, eSig int, tempDir string) func() {
	return func() {
		var matches [][]rune

		butt.SetSensitive(false)
		imgPath, err := TakeScreenshot(tempDir+string(os.PathSeparator)+"sumi", "")

		if err != nil {
			MsgBoxError(w, err.Error())
			butt.SetSensitive(true)
			return
		}

		DestroyAllChildren(&box.Container)

		label := gtk.NewLabel("Detecting...")
		box.Add(label)
		label.Show()

		go func() {
			matches, err = detectCharacters(t, imgPath)
			gdk.ThreadsEnter()
			label.SetText("")
			butt.SetSensitive(true)

			if err != nil {
				MsgBoxError(w, err.Error())
				return
			}

			boxes := generateBoxes(matches)

			DestroyAllChildren(&box.Container)

			entry.SetText("")
			entry.HandlerBlock(eSig)

			for i, e := range boxes {
				box.PackStart(e, true, true, 0)
				e.Connect("changed", cbModifyEntry(entry, i, e))
				e.Emit("changed")
			}

			entry.HandlerUnblock(eSig)
			entry.Emit("changed")

			box.ShowAll()
			gdk.Flush()
			gdk.ThreadsLeave()
		}()
	}
}

func MainWindow() {
	t, err := tesseract.NewTess("", "jpn")
	if err != nil {
		MsgBoxError(nil, err.Error())
	}

	selectButton := gtk.NewButtonWithLabel("セレクト")
	resultEntry := gtk.NewEntry()

	mainbox := gtk.NewHBox(false, 0)
	matchbox := gtk.NewHBox(false, 0)
	otherbox := gtk.NewVBox(false, 0)
	toolbar := gtk.NewVBox(true, 0)

	swin := gtk.NewScrolledWindow(nil, nil)
	swin.SetPolicy(gtk.POLICY_AUTOMATIC, gtk.POLICY_NEVER)

	w := gtk.NewWindow(gtk.WINDOW_TOPLEVEL)
	w.SetTitle("すみ")

	tempDir, err := ioutil.TempDir("", "sumi")
	if err != nil {
		MsgBoxError(w, err.Error())
	}

	swin.AddWithViewPort(matchbox)
	toolbar.PackStart(selectButton, true, true, 0)
	otherbox.PackStart(resultEntry, true, true, 0)
	otherbox.PackStart(swin, false, false, 0)

	sig := resultEntry.Connect("changed", func() {
		fmt.Println(resultEntry.GetText())
	})
	selectButton.Connect("clicked", cbSelectArea(w, t, selectButton, &matchbox.Box, resultEntry, sig, tempDir))

	mainbox.PackStart(toolbar, false, false, 0)
	mainbox.PackStart(otherbox, true, true, 0)

	w.Add(mainbox)
	w.Connect("destroy", func() {
		t.Close()
		cleanup(tempDir)
		gtk.MainQuit()
	})

	w.ShowAll()
}

func main() {
	glib.ThreadInit()
	gdk.ThreadsInit()
	gtk.Init(nil)
	gdk.ThreadsEnter()
	MainWindow()
	gtk.Main()
	gdk.ThreadsLeave()
}
