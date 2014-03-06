package main

import (
	"fmt"
    glfw "github.com/go-gl/glfw3"
)

func monitorResolution() (width, height int, err error) {
	monitor, err := glfw.GetPrimaryMonitor()
	if err != nil {
		err = fmt.Errorf("Failed to find primary monitor: %v\n", err)
		return
	}

	resolution, err := monitor.GetVideoMode()
	if err != nil {
		err = fmt.Errorf("Failed to discover video mode: %v\n", err)
		return
	}
	width = resolution.Width
	height = resolution.Height
	return
}

func openWindow(title string, width, height int) (*glfw.Window, error) {
	
    window, err := glfw.CreateWindow(width, height, title, nil, nil)
    if err != nil {
		return nil, err
    }

    window.MakeContextCurrent()
    
	//enable vertical sync (must be after MakeCurrentContext)
    glfw.SwapInterval(1)
	
	return window, nil
}