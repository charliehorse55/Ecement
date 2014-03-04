package main

import (
    glfw "github.com/go-gl/glfw3"
)

var keys []glfw.Key

type keyboard struct {
	w *glfw.Window
} 

func (k *keyboard)Begin(w *glfw.Window, num int) error {
	k.w = w
	keys = []glfw.Key{glfw.Key1, glfw.Key2, glfw.Key3, glfw.Key4, glfw.Key5, glfw.Key6, glfw.Key7, glfw.Key8, glfw.Key9, glfw.Key0}
	return nil
}

func (k *keyboard)Update(intensity []float32) error {
	for i := range intensity {
		if k.w.GetKey(keys[i]) == glfw.Press {
			intensity[i] = 1.0
		} else {
			intensity[i] = 0.0
		}
	}
	return nil
}


