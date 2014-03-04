package main

import (
    glfw "github.com/go-gl/glfw3"
)

type keyScroll struct {
	currIntensity []float64
	currSelected int
} 


func (k *keyScroll)didScroll(w *glfw.Window, xoff float64, yoff float64) {
	// _, height := w.GetSize()
	k.currIntensity[k.currSelected] -= yoff/100.0
	if k.currIntensity[k.currSelected] > 1.0 {
		k.currIntensity[k.currSelected] = 1.0
	} else if k.currIntensity[k.currSelected] < 0.0 {		
		k.currIntensity[k.currSelected] = 0.0
	}
}

func (k *keyScroll)keyPress(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if w.GetAttribute(glfw.Focused) == glfw.True && action == glfw.Release {
		var index int;
		switch key {
		case glfw.Key1:
			index = 0
		case glfw.Key2:
			index = 1
		case glfw.Key3:
			index = 2
		case glfw.Key4:
			index = 3
		case glfw.Key5:
			index = 4
		case glfw.Key6:
			index = 5
		case glfw.Key7:
			index = 6
		case glfw.Key8:
			index = 7
		case glfw.Key9:
			index = 8
		case glfw.Key0:
			index = 9
		}
		if index < len(k.currIntensity) {
			k.currSelected = index
		}
	}
}

func (k *keyScroll)Begin(w *glfw.Window, num int) error {
	w.SetScrollCallback(k.didScroll)
	w.SetKeyCallback(k.keyPress)
	k.currIntensity = make([]float64, num)
	return nil
}

func (k *keyScroll)Update(intensity []float32) error {
	for i := range intensity {
		intensity[i] = float32(k.currIntensity[i])
	}
	return nil
}


