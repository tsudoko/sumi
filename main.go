package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	"github.com/tsudoko/go.tesseract"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

var lang = flag.String("l", "jpn", "language(s) used for OCR")

func handleSignals(c chan os.Signal, w *gtk.Window) {
	select {
	case <-c:
		glib.IdleAdd(func() {
			w.Emit("destroy")
		})
	}
}

func generateBoxes(matches [][]rune) []*gtk.ComboBoxText {
	boxes := make([]*gtk.ComboBoxText, 0, 3)
	for i, b := range matches {
		cb, err := gtk.ComboBoxTextNew()
		if err != nil {
			MsgBoxError(nil, fmt.Sprintf("error generating the box %d: %s", i, err.Error()))
		}

		for _, m := range b {
			cb.AppendText(string(m))
		}
		cb.SetActive(0)
		boxes = append(boxes, cb)
	}
	return boxes
}

func cbTerminate(t *tesseract.Tess, path string) func() {
	return func() {
		t.Close()
		os.RemoveAll(path)
		gtk.MainQuit()
	}
}

func cbModifyEntry(e *gtk.Entry, i int, cbt *gtk.ComboBoxText) func() {
	return func() {
		old, err := e.GetText()
		runes := []rune(old)

		if err != nil {
			MsgBoxError(nil, "error getting text from the entry: "+err.Error())
			return
		}

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

func cbSelectArea(w *gtk.Window, t *tesseract.Tess, butt *gtk.Button, box *gtk.Box, entry *gtk.Entry, eSig glib.SignalHandle, tempDir string) func() {
	return func() {
		var matches [][]rune

		butt.SetSensitive(false)
		imgPath, err := TakeScreenshot(tempDir+string(os.PathSeparator)+"sumi", os.Getenv("SUMI_SCREENCAPTURE"))

		if err != nil {
			MsgBoxError(w, "error taking screenshot: "+err.Error())
			butt.SetSensitive(true)
			return
		}

		DestroyAllChildren(&box.Container)

		label, err := gtk.LabelNew("Detecting...")
		if err != nil {
			MsgBoxError(w, "error creating the label: "+err.Error())
			butt.SetSensitive(true)
			return
		}

		box.Add(label)
		label.Show()

		go func() {
			matches, err = detectCharacters(t, imgPath)
			glib.IdleAdd(func() {
				label.SetText("")
				butt.SetSensitive(true)

				if err != nil {
					MsgBoxError(w, "error detecting characters: "+err.Error())
					return
				}

				boxes := generateBoxes(matches)

				DestroyAllChildren(&box.Container)

				entry.SetText("")
				entry.HandlerBlock(eSig)

				for i, e := range boxes {
					box.PackStart(e, true, true, 0)
					_, err = e.Connect("changed", cbModifyEntry(entry, i, e))
					if err != nil {
						MsgBoxError(w, fmt.Sprintf("error connecting the `changed' signal to box %d: %s", i, err.Error()))
					}

					e.Emit("changed")
				}

				entry.HandlerUnblock(eSig)
				entry.Emit("changed")

				box.ShowAll()
			})
		}()
	}
}

func MainWindow(t *tesseract.Tess, tempDir string) (*gtk.Window, error) {
	selectButton, err := gtk.ButtonNewWithLabel("セレクト")
	if err != nil {
		return nil, errors.New("error creating the select button: " + err.Error())
	}

	resultEntry, err := gtk.EntryNew()
	if err != nil {
		return nil, errors.New("error creating the entry: " + err.Error())
	}

	mainbox, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
	if err != nil {
		return nil, errors.New("error creating the mainbox: " + err.Error())
	}

	matchbox, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
	if err != nil {
		return nil, errors.New("error creating the matchbox: " + err.Error())
	}

	ocrbox, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
	if err != nil {
		return nil, errors.New("error creating the ocrbox: " + err.Error())
	}

	swin, err := gtk.ScrolledWindowNew(nil, nil)
	if err != nil {
		return nil, errors.New("error creating the scrolled window: " + err.Error())
	}

	w, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		return nil, errors.New("error creating the window: " + err.Error())
	}

	w.SetTitle("すみ")
	matchbox.SetHomogeneous(true)
	swin.SetOverlayScrolling(false)
	swin.SetPolicy(gtk.POLICY_AUTOMATIC, gtk.POLICY_NEVER)

	sig, err := resultEntry.Connect("changed", func() {
		text, err := resultEntry.GetText()
		if err != nil {
			MsgBoxError(nil, "error getting text from the entry: "+err.Error())
			return
		}
		fmt.Println(text)
	})
	if err != nil {
		return w, errors.New("error connecting the `changed' signal to the entry: " + err.Error())
	}

	_, err = selectButton.Connect("clicked", cbSelectArea(w, t, selectButton, matchbox, resultEntry, sig, tempDir))
	if err != nil {
		return w, errors.New("error connecting the `clicked' signal to the select button: " + err.Error())
	}

	_, err = w.Connect("destroy", cbTerminate(t, tempDir))
	if err != nil {
		return w, errors.New("error connecting the `destroy' signal to the window: " + err.Error())
	}

	swin.Add(matchbox)
	ocrbox.PackStart(resultEntry, false, false, 0)
	ocrbox.PackStart(swin, true, true, 0)
	mainbox.PackStart(selectButton, false, false, 0)
	mainbox.PackStart(ocrbox, true, true, 0)
	w.Add(mainbox)

	w.ShowAll()

	return w, nil
}

func main() {
	gtk.Init(&os.Args)
	flag.Parse()

	t, err := tesseract.NewTess("", *lang)
	if err != nil {
		MsgBoxError(nil, "error initializing tesseract: "+err.Error())
		return
	}

	tempDir, err := ioutil.TempDir("", "sumi")
	if err != nil {
		MsgBoxError(nil, "error creating the temporary directory: "+err.Error())
		return
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	w, err := MainWindow(t, tempDir)
	if err != nil {
		MsgBoxError(nil, err.Error())
		return
	}

	go handleSignals(c, w)
	gtk.Main()
}
