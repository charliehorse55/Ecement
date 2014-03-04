package main

import (
	"flag"
	"io/ioutil"
	"fmt"
	"os"
	"log"
	"runtime"
	"time"
    glfw "github.com/go-gl/glfw3"
    gl "github.com/go-gl/gl"
)

var channel1 float64

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

func didScroll(w *glfw.Window, xoff float64, yoff float64) {
	// _, height := w.GetSize()
	channel1 -= yoff/100
	if channel1 > 1.0 {
		channel1 = 1.0
	} else if channel1 < 0.0 {		
		channel1 = 0.0
	}
}

func main() {
	flag.Parse()
	
	imagePaths := flag.Args()
	
	if len(imagePaths) < 2 {
		log.Printf("Ecement requires at least 2 input files")
		return
	}	
	
	glfw.SetErrorCallback(errorCallback)

    if !glfw.Init() {
        return
    }
    defer glfw.Terminate()

	glfw.WindowHint(glfw.OpenglForwardCompatible, glfw.True)
	glfw.WindowHint(glfw.OpenglProfile, glfw.OpenglCoreProfile)
	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 2)

    window, err := glfw.CreateWindow(640, 480, "Testing", nil, nil)
    if err != nil {
        panic(err)
    }

    window.MakeContextCurrent()
	window.SetScrollCallback(didScroll)
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
	
	tex1location := shaderProgram.GetUniformLocation("tex1")
	tex1location.Uniform1i(1)
	checkGLError()	
	
	intensitylocation := shaderProgram.GetUniformLocation("intensity")
	intensitylocation.Uniform2f(0.5, 0.5)
	checkGLError()	
	
	
	//load the textures
	for i,filename := range imagePaths {
		file, err := os.Open(filename)
		if err != nil {
			log.Printf("Failed to open %s: %v", filename, err)
			continue
		}
		defer file.Close()
		
		_, err = loadImage(file, i)
		if err != nil {
			log.Printf("Failed to load texture: %v", err)
		}
	}
	checkGLError()	

	lastUpdated := time.Now()
	frames := 0
    for !window.ShouldClose() {
		frames++
				
		//set the intensities
		intensitylocation.Uniform2f(float32(channel1), 0.5)		
		
		gl.DrawElements(gl.TRIANGLES, 6, gl.UNSIGNED_INT, nil)
		checkGLError()
		
		
        glfw.PollEvents()
		now := time.Now()
		diff := now.Sub(lastUpdated)
		if diff > time.Second {
			fmt.Printf("%.2f FPS\n", float64(frames)/diff.Seconds())
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