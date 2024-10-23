package main

import "github.com/AllenDang/gform"

func main() {
	gform.Init()

	mainForm := gform.NewForm(nil)
	mainForm.SetPos(300, 500)
	mainForm.SetSize(800, 500)
	mainForm.SetCaption("Lab 7")

	mainForm.Show()
	gform.RunMainLoop()
}
