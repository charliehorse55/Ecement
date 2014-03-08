package main

import (
	"flag"
	"io/ioutil"
	"fmt"
	"log"
	"runtime"
	"time"
    glfw "github.com/go-gl/glfw3"
    gl "github.com/go-gl/gl"
)

type Pixel struct {
	R, G, B, A float32
}

type intensityController interface {
	Begin(w *glfw.Window, num int) error
	Update(intensity []float32) error
}

var controller intensityController
var currIntensity []float32


// OpenGL and glfw need to be called from the main thread
func init() {
    runtime.LockOSThread()
}

func checkGLError() {
	glErr := gl.GetError()
	if glErr != 0 {
		log.Printf("GL Error Code: %d\n", int(glErr))
		panic("stack trace")
	}
}

func loadShader(filename string, kind gl.GLenum) (gl.Shader, error)  {
	source, err := ioutil.ReadFile(filename)
	if err != nil {
		return 0, err
	}
	
	newShader := gl.CreateShader(kind)
	newShader.Source(string(source))
	newShader.Compile()
	compileStatus := newShader.Get(gl.COMPILE_STATUS)
	if compileStatus != gl.TRUE {
		return 0, fmt.Errorf("Compilation failed:\n%s", newShader.GetInfoLog())
	}
	return newShader, nil
}

func errorCallback(err glfw.ErrorCode, desc string) {
    log.Fatalf("%v: %v\n", err, desc)
}

func main() {
	flag.Parse()
	
	imagePaths := flag.Args()
	
	if len(imagePaths) < 2 || len(imagePaths) > 10 {
		log.Fatalf("Ecement requires n files, where 2 <= n <= 10\n")
	}	
	
	width, height, err := getImageRes(imagePaths[0])
	if err != nil {
		log.Fatalf("Failed to read image: %v\n", err)
	}
	
	controller = &keyScroll{}
	// controller = &keyboard{}
	// controller = &sincos{}
	
	glfw.SetErrorCallback(errorCallback)
    if !glfw.Init() {
        return
    }
    defer glfw.Terminate()

	glfw.WindowHint(glfw.OpenglForwardCompatible, glfw.True)
	glfw.WindowHint(glfw.OpenglProfile, glfw.OpenglCoreProfile)
	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 2)
	
	glfw.WindowHint(glfw.Resizable, glfw.False)

	_, monitorHeight, err := monitorResolution()
	if err != nil {
		log.Fatalf("Failed to discover monitor resolution: %v", err)
	}
	
	windowHeight := int(0.8 * float64(monitorHeight))
	windowWidth := int(float64(width)/float64(height) * float64(windowHeight))

	window, err := openWindow("Ecement", windowWidth, windowHeight)
	if err != nil {
		log.Fatalf("Failed to open window: %v", err)
	}

	result := gl.Init()
	if result != 0 {
		log.Fatalf("Failed to initialize GLEW: %d", result)
	}
	gl.GetError()


	err = controller.Begin(window, len(imagePaths))
	if err != nil {
		log.Fatalf("Failed to initialze controller: %v", err)
	}
			
	VAO := gl.GenVertexArray()
	VAO.Bind()

	//just drawing a rectangle
	vertices := []float32{
		0.0, 1.0,	  // Top-left
		1.0, 1.0,     // Top-right
		1.0, 0.0,     // Bottom-right
		0.0, 0.0,     // Bottom-left
	}

	elements := []uint32{
		0, 1, 2,
		2, 3, 0,
	}

	vertexBuf := gl.GenBuffer()
	vertexBuf.Bind(gl.ARRAY_BUFFER)
	gl.BufferData(gl.ARRAY_BUFFER, 4*len(vertices), vertices, gl.STATIC_DRAW)

	elementBuf := gl.GenBuffer()
	elementBuf.Bind(gl.ELEMENT_ARRAY_BUFFER)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, 4*len(elements), elements, gl.STATIC_DRAW)
	checkGLError()

	vertexShader, err := loadShader("vertex.glsl", gl.VERTEX_SHADER)
	if err != nil {
		log.Printf("Failed to load vertex shader: %v", err)
		return
	}
	fragmentShader, err := loadShader("fragment.glsl", gl.FRAGMENT_SHADER)
	if err != nil {
		log.Printf("Failed to load fragment step shader: %v", err)
		return
	}
	
	finishShader, err := loadShader("finish.glsl", gl.FRAGMENT_SHADER)
	if err != nil {
		log.Printf("Failed to load fragment finishing shader: %v", err)
		return
	}
	
	
	//create a program to iteratively cement each image
	stepProgram := gl.CreateProgram()
	stepProgram.AttachShader(vertexShader)
	stepProgram.AttachShader(fragmentShader)
	checkGLError()
		
	stepProgram.Link()
	checkGLError()
	stepProgram.Use()
	
	//create a program to apply the final tonemap to generate a viewable image
	finishProgram := gl.CreateProgram()
	finishProgram.AttachShader(vertexShader)
	finishProgram.AttachShader(finishShader)
	checkGLError()
		
	finishProgram.Link()
	checkGLError()
		
	posLocation := stepProgram.GetAttribLocation("position")
	posLocation.AttribPointer(2, gl.FLOAT, false, 0, nil)
	posLocation.EnableArray()
	checkGLError()	
	
	oldlocation := stepProgram.GetUniformLocation("old")
	oldlocation.Uniform1i(0)
	checkGLError()	
	
	newlocation := stepProgram.GetUniformLocation("new")
	newlocation.Uniform1i(1)
	checkGLError()	
		
	intensitylocation := stepProgram.GetUniformLocation("intensity")
	checkGLError()	
			
	//load the textures
	vectors, err := loadImages(imagePaths, width, height)
	if err != nil {
		log.Fatalf("Failed to load images into texture: %v", err)
	}
	checkGLError()	

	//create framebuffers to render final outputs	
	fullSizeFBs := createFramebufferPair(width, height)
	
	//create framebuffers to render in real time for the current window
	windowFBs := createFramebufferPair(windowWidth, windowHeight)
		
	checkGLError()
	
	screenFB := gl.Framebuffer(0)
				
	lastUpdated := time.Now()
	frames := 0
	shouldSave := false
	rawOutput := make([]Pixel, width*height)
	done := make(chan int, 100)
	saveOperations := 0
    for !window.ShouldClose() {
		frames++
					
		err = controller.Update(currIntensity)
		if err != nil {
			log.Printf("Failed to update controller state: %v", err)
			return
		}
		
		//render to the screen
		stepProgram.Use()
		gl.Viewport(0,0, windowWidth, windowHeight)
		render(vectors, windowFBs, intensitylocation)
		
		screenFB.Bind()
		finishProgram.Use()
		gl.DrawElements(gl.TRIANGLES, 6, gl.UNSIGNED_INT, nil)
		
		checkGLError()		
	
		//save the output if the user wanted to
		if shouldSave {
			// windowFB.Bind()
			// checkGLError()	
			stepProgram.Use()
			gl.Viewport(0,0, width, height)
			render(vectors, fullSizeFBs, intensitylocation)
			
			finishProgram.Use()
			gl.DrawElements(gl.TRIANGLES, 6, gl.UNSIGNED_INT, nil)
			
			gl.ReadPixels(0, 0, width, height, gl.RGBA, gl.FLOAT, rawOutput)
		
			go func() {
				path := "output.jpg"
				err := saveToJPEG(path, width, height, rawOutput)
				if err != nil {
					log.Printf("Failed to save jpeg: %v", err)
				}
				done <- 1
			}()
			saveOperations++
			shouldSave = false	
		}
				
		glfw.PollEvents()
		now := time.Now()
		diff := now.Sub(lastUpdated)
		if diff > time.Second*2 {
			fmt.Printf("%.1f FPS\n", float64(frames)/diff.Seconds())
			lastUpdated = now
			frames = 0
		}
		
		window.SwapBuffers()
    }
	
	//wait for any remaining saves to complete
	for i := 0; i < saveOperations; i++ {
		<-done
	}	
}