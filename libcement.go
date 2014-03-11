package libcement

import (
	"fmt"
	"log"
    gl "github.com/go-gl/gl"
)

type RGB32f struct {
	R, G, B float32
}

type Lightvector struct {
	Texture gl.Texture
	Intensity RGB32f
	Filename string
}

type Painting []Lightvector

type Rendering struct {
	Painting Painting
	curr []RGB32f
	Texture gl.Texture
	BackBuffer gl.Texture
	Width, Height int
}

func (p Painting)Load(removeBackground bool) error {
	FB.Bind()
	vao.Bind()
	preprocess.Use()
	if len(p) == 0 {
		return fmt.Errorf("Empty painting")
	}
	
	width, height, err := getImageRes(p[0].Filename)
	if err != nil {
		return err
	}
	gl.Viewport(0,0, width, height)
	
	
	return loadImages(p, width, height, removeBackground)
}

func (p Painting)CreateRendering(width, height int) *Rendering {
	var result Rendering
	result.Painting = p
	result.curr = make([]RGB32f, len(p))
	result.Width = width
	result.Height = height
	
	tex := []*gl.Texture{&result.Texture, &result.BackBuffer}
	for i := range tex {
		*(tex[i]) = gl.GenTexture()
		(*(tex[i])).Bind(gl.TEXTURE_2D)
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA32F, width, height, 0, gl.RGBA, gl.FLOAT, nil)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	}
	
	//clear the current texture to black
	FB.Bind()
	vao.Bind()
	gl.Viewport(0,0, width, height)
 	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, result.Texture, 0)	
	gl.ClearColor(0.0, 0.0, 0.0, 1.0)
    gl.Clear(gl.COLOR_BUFFER_BIT)
	
	//this will render everything, as result.curr has been init to 0
	//thus the current state is that the texture contains only black
	result.Update()
	return &result
}

func (r *Rendering)Update() {
	FB.Bind()
	step.Use()
	vao.Bind()
	gl.Viewport(0,0, r.Width, r.Height)
	
	//update the rendering by adding only the changes of each channel
	//this speeds things up a lot, as usually only a few channels are 
	//changed per frame
	for i := range r.Painting {
		//does not need to be updated
		if r.Painting[i].Intensity == r.curr[i] {
			continue
		}
		
		diff := RGB32f {
			R:r.Painting[i].Intensity.R - r.curr[i].R,
			G:r.Painting[i].Intensity.G - r.curr[i].G,
			B:r.Painting[i].Intensity.B - r.curr[i].B,
		} 
		r.curr[i] = r.Painting[i].Intensity
				
		//update the uniform to use the new intensities
		intensitylocation.Uniform3f(diff.R, diff.G, diff.B)
	
		//bind the current result framebuffer to texture 0
		gl.ActiveTexture(gl.TEXTURE0)
		r.Texture.Bind(gl.TEXTURE_2D)
	
		//bind the input image to texture 1
		gl.ActiveTexture(gl.TEXTURE0 + gl.GLenum(1))
		r.Painting[i].Texture.Bind(gl.TEXTURE_2D)
	
		//draw to the output buffer
	 	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, r.BackBuffer, 0)
		gl.DrawElements(gl.TRIANGLES, 6, gl.UNSIGNED_INT, nil)
		
		//swap the buffers
		tmp := r.BackBuffer
		r.BackBuffer = r.Texture
		r.Texture = tmp
	}
}

//NOTE that this function specifically does NOT bind to a framebuffer
// or set a viewport
//this lets you bind to the screen before calling this if you want
//otherwise, just bind to Ecement.FB
func (r *Rendering)Tonemap() {
	vao.Bind()
	final.Use()
	gl.DrawElements(gl.TRIANGLES, 6, gl.UNSIGNED_INT, nil)	
}

func checkGLError() {
	glErr := gl.GetError()
	if glErr != 0 {
		log.Printf("GL Error Code: %d\n", int(glErr))
		panic("stack trace")
	}
}


var FB gl.Framebuffer

//bind to this before we draw
var vao gl.VertexArray

//shaders
var preprocess gl.Program
var step gl.Program
var final gl.Program

var intensitylocation gl.UniformLocation

func Start() {
	
	FB = gl.GenFramebuffer()
	FB.Bind()
	
	loadedShaders = make(map[string]gl.Shader)
		
	//just drawing a rectangle
	vao = gl.GenVertexArray()
	vao.Bind()
	vertices := []float32{
		0.0, 1.0,	  // Top-left
		1.0, 1.0,     // Top-right
		1.0, 0.0,     // Bottom-right
		0.0, 0.0,     // Bottom-left
	}
	vertexBuf := gl.GenBuffer()
	vertexBuf.Bind(gl.ARRAY_BUFFER)
	gl.BufferData(gl.ARRAY_BUFFER, 4*len(vertices), vertices, gl.STATIC_DRAW)
	
	elements := []uint32{
		0, 1, 2,
		2, 3, 0,
	}
	elementBuf := gl.GenBuffer()
	elementBuf.Bind(gl.ELEMENT_ARRAY_BUFFER)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, 4*len(elements), elements, gl.STATIC_DRAW)
	
	
	//create a program to preprocess the images (apply F inverse then subtract base image)
	var err error
	preprocess, err = loadProgram("vertex.glsl", "preprocess.glsl")
	if err != nil {
		log.Fatalf("Failed to load preprocessing program: %v", err)
	}
	
	//create a program to cement the images together (sum their pixel scaled by intensity factors)
	step, err = loadProgram("vertex.glsl", "cement.glsl")
	if err != nil {
		log.Fatalf("Failed to load cement program: %v", err)
	}
	
	//create a program to tone map the result by applying F to the cemented image
	final, err = loadProgram("vertex.glsl", "tonemap.glsl")
	if err != nil {
		log.Fatalf("Failed to load tonemap program: %v", err)
	}
	
	//set up the shader programs
	preprocess.Use()
	checkGLError()	
		
	posLocation := preprocess.GetAttribLocation("position")
	checkGLError()	
	posLocation.AttribPointer(2, gl.FLOAT, false, 0, nil)
	checkGLError()	
	posLocation.EnableArray()
	checkGLError()	
	
	oldlocation := preprocess.GetUniformLocation("img")
	oldlocation.Uniform1i(0)
	
	newlocation := preprocess.GetUniformLocation("base")
	newlocation.Uniform1i(1)
	checkGLError()	
		
	step.Use()
	posLocation = step.GetAttribLocation("position")
	posLocation.AttribPointer(2, gl.FLOAT, false, 0, nil)
	posLocation.EnableArray()
	checkGLError()	
	
	totalLocation := step.GetUniformLocation("total")
	totalLocation.Uniform1i(0)
	checkGLError()	
	
	imglocation := step.GetUniformLocation("img")
	imglocation.Uniform1i(1)
	checkGLError()	
	
	intensitylocation = step.GetUniformLocation("intensity")
	checkGLError()
	
	final.Use()
	posLocation = final.GetAttribLocation("position")
	posLocation.AttribPointer(2, gl.FLOAT, false, 0, nil)
	posLocation.EnableArray()
	checkGLError()	
	
	imglocation = final.GetUniformLocation("img")
	imglocation.Uniform1i(0)
	checkGLError()	
}
