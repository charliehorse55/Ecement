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

var currIntensity []float32

type intensityController interface {
	Begin(w *glfw.Window, num int) error
	Update(intensity []float32) error
}

var controller intensityController


// OpenGL and glfw need to be called from the main thread
func init() {
    runtime.LockOSThread()
}

func checkGLError() {
	glErr := gl.GetError()
	if glErr != 0 {
		fmt.Printf("Code: %d\n", glErr)
		panic("gl error")
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
    fmt.Printf("%v: %v\n", err, desc)
}

func main() {
	flag.Parse()
	
	imagePaths := flag.Args()
	
	if len(imagePaths) < 2 || len(imagePaths) > 10 {
		log.Printf("Ecement requires n files, where 2 <= n <= 10\n")
		return
	}	
	
	Width, Height, err := getImageRes(imagePaths[0])
	if err != nil {
		log.Printf("Failed to read image: %v\n", err)
		return
	}
	
	// controller = &keyScroll{}
	controller = &keyboard{}
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

	monitor, err := glfw.GetPrimaryMonitor()
	if err != nil {
		log.Printf("Failed to find primary monitor: %v\n", err)
		return
	}

	resolution, err := monitor.GetVideoMode()
	if err != nil {
		log.Printf("Failed to discover video mode: %v\n", err)
		return
	}
	
	resolution.Height = int(0.8 * float64(resolution.Height))
	resolution.Width = int(float64(Width)/float64(Height) * float64(resolution.Height))

    window, err := glfw.CreateWindow(resolution.Width, resolution.Height, "Testing", nil, nil)
    if err != nil {
		log.Printf("Ecement requires n files, where 2 <= n <= 10\n")
		return
    }

    window.MakeContextCurrent()
	
	err = controller.Begin(window, len(imagePaths))
	if err != nil {
		log.Printf("Failed to initialze controller: %v", err)
		return
	}
			
	gl.Init()
	gl.GetError()
	
    //enable vertical sync (must be after MakeCurrentContext)
    glfw.SwapInterval(1)

	VAO := gl.GenVertexArray()
	VAO.Bind()

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
	err = loadImages(imagePaths, Width, Height)
	if err != nil {
		log.Printf("Failed to load images into texture: %v", err)
		return
	}
	checkGLError()	

	lastUpdated := time.Now()
	frames := 0
    for !window.ShouldClose() {
		frames++
				
		//set the intensities
		err = controller.Update(currIntensity)
		if err != nil {
			log.Printf("Failed to update controller state: %v", err)
			return
		}
		
		intensitylocation.Uniform1fv(len(currIntensity), currIntensity)
		
		gl.DrawElements(gl.TRIANGLES, 6, gl.UNSIGNED_INT, nil)
		checkGLError()
		
		
        glfw.PollEvents()
		now := time.Now()
		diff := now.Sub(lastUpdated)
		if diff > time.Second*10 {
			fmt.Printf("%.1f FPS\n", float64(frames)/diff.Seconds())
			lastUpdated = now
			frames = 0
		}
		
        window.SwapBuffers()
    }
					
	// outfilepath := "output.jpg"
	// outfile, err := os.Create(outfilepath)
	// if err != nil {
	// 	log.Printf("Failed to open output file: %s", outfilepath)
	// 	return 
	// }
	// defer outfile.Close()
	// 
	// err = jpeg.Encode(outfile, output, nil)
	// if err != nil {
	// 	log.Printf("Failed to encode output file")
	// }
}