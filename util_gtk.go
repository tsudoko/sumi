package main

import "github.com/gotk3/gotk3/gtk"

func MsgBoxError(w *gtk.Window, msg string) {
	d := gtk.MessageDialogNew(w, 0, gtk.MESSAGE_ERROR, gtk.BUTTONS_OK, msg)
	d.Run()
	d.Destroy()
}

func DestroyAllChildren(c *gtk.Container) {
	children := c.GetChildren()
	children.Foreach(func(w interface{}) {
		//child, ok := w.(*gtk.Widget)
		if child, ok := w.(*gtk.Widget); ok {
			child.Destroy()
		}
	})
	children.Free()
}
