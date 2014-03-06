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
		log.Fatal("GL Error Code: %d\n", glErr)
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
    log.Fatal("%v: %v\n", err, desc)
}

func main() {
	flag.Parse()
	
	imagePaths := flag.Args()
	
	if len(imagePaths) < 2 || len(imagePaths) > 10 {
		log.Fatal("Ecement requires n files, where 2 <= n <= 10\n")
	}	
	
	width, height, err := getImageRes(imagePaths[0])
	if err != nil {
		log.Fatal("Failed to read image: %v\n", err)
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

	_, monitorHeight, err := monitorResolution()
	if err != nil {
		log.Fatal("Failed to discover monitor resolution: %v", err)
	}
	
	windowHeight := int(0.8 * float64(monitorHeight))
	windowWidth := int(float64(width)/float64(height) * float64(windowHeight))

	window, err := openWindow("Ecement", windowWidth, windowHeight)
	if err != nil {
		log.Fatal("Failed to open window: %v", err)
	}

	result := gl.Init()
	if result != 0 {
		log.Fatal("Failed to initialize GLEW: %d", result)
	}
	gl.GetError()


	err = controller.Begin(window, len(imagePaths))
	if err != nil {
		log.Fatal("Failed to initialze controller: %v", err)
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
		log.Printf("Failed to load fragment shader: %v", err)
		return
	}
	
	shaderProgram := gl.CreateProgram()
	shaderProgram.AttachShader(vertexShader)
	shaderProgram.AttachShader(fragmentShader)
	checkGLError()
	
	shaderProgram.Link()
	checkGLError()
	shaderProgram.Use()
	checkGLError()
	
	posLocation := shaderProgram.GetAttribLocation("position")
	posLocation.AttribPointer(2, gl.FLOAT, false, 0, nil)
	posLocation.EnableArray()
	checkGLError()	
	
	tex0location := shaderProgram.GetUniformLocation("tex0")
	tex0location.Uniform1i(0)
	
	numlocation := shaderProgram.GetUniformLocation("num")
	numlocation.Uniform1i(len(imagePaths))
	checkGLError()	

	currIntensity = make([]float32, len(imagePaths))
	
	intensitylocation := shaderProgram.GetUniformLocation("intensity")
	intensitylocation.Uniform1fv(len(currIntensity), currIntensity)
	checkGLError()	
	
		
	//load the textures
	err = loadImages(imagePaths, width, height)
	if err != nil {
		log.Printf("Failed to load images into texture: %v", err)
		return
	}
	checkGLError()	

	//create a framebuffer to render into as an intermediary
	fbo := gl.GenFramebuffer()
	fbo.Bind()
	
	screen := gl.Framebuffer(0)
	
	// The texture we're going to render to
	 
	gl.ActiveTexture(gl.TEXTURE0 + gl.GLenum(1))
	 
	renderedTexture := gl.GenTexture()
	
	// "Bind" the newly created texture : all future texture functions will modify this texture
	renderedTexture.Bind(gl.TEXTURE_2D)
	
	// Give an empty image to OpenGL ( the last "0" )
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA32F, width, height, 0, gl.RGBA, gl.FLOAT, nil)
	
	// Poor filtering. Needed !
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)

 	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, renderedTexture, 0)
	gl.DrawBuffer(gl.COLOR_ATTACHMENT0)
		
	lastUpdated := time.Now()
	frames := 0
	shouldSave := false
	rawOutput := make([]Pixel, width*height)
    for !window.ShouldClose() {
		frames++
				
		//set the intensities
		// currIntensity[0] = 1.0
		// currIntensity[1] = 1.0
		// currIntensity[2] = 1.0
		
		err = controller.Update(currIntensity)
		if err != nil {
			log.Printf("Failed to update controller state: %v", err)
			return
		}
		
		intensitylocation.Uniform1fv(len(currIntensity), currIntensity)
		
		//render to the screen
		screen.Bind()
		gl.Viewport(0,0, windowWidth, windowHeight)
		gl.DrawElements(gl.TRIANGLES, 6, gl.UNSIGNED_INT, nil)
		
		//save the output if the user wanted to
		if shouldSave {
			fbo.Bind()
			gl.Viewport(0,0, width, height)
			gl.DrawElements(gl.TRIANGLES, 6, gl.UNSIGNED_INT, nil)
			gl.GetTexImage(gl.TEXTURE_2D, 0, gl.RGBA, gl.FLOAT, rawOutput)
			go saveToJPEG("output.jpg", width, height, rawOutput)
			shouldSave = false
		}
				
		glfw.PollEvents()
		now := time.Now()
		diff := now.Sub(lastUpdated)
		if diff > time.Second*10 {
			fmt.Printf("%.1f FPS\n", float64(frames)/diff.Seconds())
			lastUpdated = now
			frames = 0
			shouldSave = true
		}
		
		window.SwapBuffers()
    }
	
	//print the height/width of the target texture
	
}