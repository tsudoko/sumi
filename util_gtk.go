package main

import (
	"unsafe"

	"github.com/mattn/go-gtk/gtk"
)

func MsgBoxError(w *gtk.Window, msg string) {
	d := gtk.NewMessageDialog(w, 0, gtk.MESSAGE_ERROR, gtk.BUTTONS_OK, msg)
	d.Run()
	d.Destroy()
}

func DestroyAllChildren(c *gtk.Container) {
	children := c.GetChildren()
	children.ForEach(func(p unsafe.Pointer, _ interface{}) {
		child := gtk.WidgetFromNative(p)
		child.Destroy()
	})
	children.Free()
}
