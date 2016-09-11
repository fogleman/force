package main

import (
	"math/rand"
	"runtime"
	"time"

	"github.com/fogleman/force"
	"github.com/fogleman/imview"
	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
)

func init() {
	runtime.LockOSThread()
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	if err := gl.Init(); err != nil {
		panic(err)
	}
	if err := glfw.Init(); err != nil {
		panic(err)
	}
	defer glfw.Terminate()

	model := force.NewModel()

	start := time.Now()
	previous := 0.0

	window, err := imview.NewWindow(model.Draw(1024, 1024, 0, 0))
	if err != nil {
		panic(err)
	}

	for {
		if window.ShouldClose() {
			window.Destroy()
			break
		}
		t := time.Since(start).Seconds()
		dt := t - previous
		previous = t
		model.Step(t, dt)
		x, y := window.GetCursorPos()
		im := model.Draw(1600, 1600, x, y)
		window.SetImage(im)
		glfw.PollEvents()
	}
}
