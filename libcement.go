package libcement

import (
	"log"
	"fmt"
    gl "github.com/go-gl/gl"
)

const (
	internalFormat = gl.RGB32F
)

type RGB32f struct {
	R, G, B float32
}

type Painting struct {
	background gl.Texture
	vectors []gl.Texture
	F []float32
	Finverse []float32
	lutF gl.Texture
	lutFinverse gl.Texture
	Width int
	Height int
}

type Rendering struct {
	Painting *Painting
	curr []RGB32f
	FrontBuffer gl.Texture
	BackBuffer gl.Texture
	Width, Height int
}

func NewPainting(width, height int, Finverse []float32, background gl.Texture) *Painting {
	p := &Painting{Finverse:Finverse, Width:width, Height:height}
	if len(p.Finverse) == 0 {
		p.Finverse = createSquaredFinverse(256)
	}
	
	p.F = FinverseToF(p.Finverse)
	
	//create the LUTs
	p.lutF = createLookupTable(p.F)	
	p.lutFinverse = createLookupTable(p.Finverse)
	
	//create a 1x1 black texture to preprocess the background image with
	//this way we won't subtract anything from the background image
	smallBlackTex := gl.GenTexture()
	defer smallBlackTex.Delete()
	
	smallBlackTex.Bind(gl.TEXTURE_2D)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	pixel := []uint8{0, 0, 0, 255}
    gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, 1, 1, 0, gl.RGBA, gl.UNSIGNED_BYTE, pixel)
	
	//load the background
	p.background = loadTexture(width, height, background, smallBlackTex, p.lutFinverse)
		
	return p
}

func (p *Painting)AddLightvector(newVector gl.Texture) {
	newTexture := loadTexture(p.Width, p.Height, newVector, p.background, p.lutFinverse)
	p.vectors = append(p.vectors, newTexture)
}

func createLookupTable(data []float32) gl.Texture {
	tex := gl.GenTexture()
	tex.Bind(gl.TEXTURE_1D)
	gl.TexParameteri(gl.TEXTURE_1D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_BORDER)
	gl.TexParameteri(gl.TEXTURE_1D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_BORDER)
	gl.TexParameteri(gl.TEXTURE_1D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_1D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexImage1D(gl.TEXTURE_1D, 0, gl.RED, len(data), 0, gl.RED, gl.FLOAT, data)	
	return tex
}

func (p *Painting)CreateRendering(width, height int, initial []RGB32f) *Rendering {
	var result Rendering
	result.Painting = p
	result.curr = make([]RGB32f, len(p.vectors))
	result.Width = width
	result.Height = height
	
	tex := []*gl.Texture{&result.FrontBuffer, &result.BackBuffer}
	for i := range tex {
		*(tex[i]) = gl.GenTexture()
		(*(tex[i])).Bind(gl.TEXTURE_2D)
		gl.TexImage2D(gl.TEXTURE_2D, 0, internalFormat, width, height, 0, gl.RGB, gl.FLOAT, nil)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	}
	
	//clear the current texture to black
	FB.Bind()
	vao.Bind()
	gl.Viewport(0,0, width, height)
 	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, result.FrontBuffer, 0)	
	gl.ClearColor(0.0, 0.0, 0.0, 1.0)
    gl.Clear(gl.COLOR_BUFFER_BIT)
	
	//add the background
	step.Use()
	result.cementVector(result.Painting.background, 1.0, 1.0, 1.0)
	
	//this will render everything, as result.curr has been init to 0
	//thus the current state is that the texture contains only black
	result.Update(initial)
	return &result
}

func (r *Rendering)cementVector(tex gl.Texture, R, G, B float32) {
	
	//update the uniform to use the new intensities
	intensitylocation.Uniform3f(R, G, B)

	//bind the current result framebuffer to texture 0
	gl.ActiveTexture(gl.TEXTURE0)
	r.FrontBuffer.Bind(gl.TEXTURE_2D)

	//bind the input image to texture 1
	gl.ActiveTexture(gl.TEXTURE0 + gl.GLenum(1))
	tex.Bind(gl.TEXTURE_2D)

	//draw to the output buffer
 	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, r.BackBuffer, 0)
	gl.DrawElements(gl.TRIANGLES, 6, gl.UNSIGNED_INT, nil)
	
	//swap the buffers
	tmp := r.BackBuffer
	r.BackBuffer = r.FrontBuffer
	r.FrontBuffer = tmp
}

func loadTexture(width, height int, source, background, Finverse gl.Texture) gl.Texture {	
	FB.Bind()
	vao.Bind()
	preprocess.Use()
	gl.Viewport(0,0, width, height)
			
	//create the actual texture for image (where the preprocessed result goes)
	result := gl.GenTexture()
	result.Bind(gl.TEXTURE_2D)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexImage2D(gl.TEXTURE_2D, 0, internalFormat, width, height, 0, gl.RGBA, gl.UNSIGNED_BYTE, nil)	
	
	//bind it as the framebuffer target
 	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, result, 0)	
	
	//source is in tex unit 0
	gl.ActiveTexture(gl.TEXTURE0)
	source.Bind(gl.TEXTURE_2D)
	
	//background is in texture unit 1
	gl.ActiveTexture(gl.TEXTURE0 + gl.GLenum(1))
	background.Bind(gl.TEXTURE_2D)
				
	//lut for Finverse is in texture unit 3
	gl.ActiveTexture(gl.TEXTURE0 + gl.GLenum(3))
	Finverse.Bind(gl.TEXTURE_1D)
								
	//preprocess the image
	gl.DrawElements(gl.TRIANGLES, 6, gl.UNSIGNED_INT, nil)		
	
	return result
}	

func (r *Rendering)Update(newIntensity []RGB32f) {
	FB.Bind()
	step.Use()
	vao.Bind()
	
	gl.Viewport(0,0, r.Width, r.Height)
	
	//update the rendering by adding only the changes of each channel
	//this speeds things up a lot, as usually only a few channels are 
	//changed per frame
	for i := range r.curr {
		//does not need to be updated
		if newIntensity[i] == r.curr[i] {
			continue
		}
		
		diff := RGB32f {
			R:newIntensity[i].R - r.curr[i].R,
			G:newIntensity[i].G - r.curr[i].G,
			B:newIntensity[i].B - r.curr[i].B,
		} 
		r.curr[i] = newIntensity[i]
		
		r.cementVector(r.Painting.vectors[i], diff.R, diff.G, diff.B)
	}
}

//NOTE that this function specifically does NOT bind to a framebuffer
// or set a viewport
//this lets you bind to the screen before calling this if you want
//otherwise, just bind to libcement.FB
func (r *Rendering)Tonemap() {	
	vao.Bind()
	tonemap.Use()
	
	//the tonemap shader expects the current lightspace to be in tex unit 0
	gl.ActiveTexture(gl.TEXTURE0)
	r.FrontBuffer.Bind(gl.TEXTURE_2D)
	
	//the lut for F() needs to be in tex unit 3
	gl.ActiveTexture(gl.TEXTURE0 + gl.GLenum(3))
	r.Painting.lutF.Bind(gl.TEXTURE_1D)
	
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
var tonemap gl.Program

var intensitylocation gl.UniformLocation

func Start() error {
	
	err := ShaderInit() 
	if err != nil {
		return err
	}
	
	FB = gl.GenFramebuffer()
	FB.Bind()
			
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
	preprocess, err = loadProgram("vertex.glsl", "preprocess.glsl")
	if err != nil {
		return fmt.Errorf("Failed to load preprocessing program: %v", err)
	}
	
	//create a program to cement the images together (sum their pixel scaled by intensity factors)
	step, err = loadProgram("vertex.glsl", "cement.glsl")
	if err != nil {
		return fmt.Errorf("Failed to load cement program: %v", err)
	}
	
	//create a program to tone map the result by applying F to the cemented image
	tonemap, err = loadProgram("vertex.glsl", "tonemap.glsl")
	if err != nil {
		return fmt.Errorf("Failed to load tonemap program: %v", err)
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
	
	Finverselocation := preprocess.GetUniformLocation("Finverse")
	Finverselocation.Uniform1i(3)
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
	
	tonemap.Use()
	posLocation = tonemap.GetAttribLocation("position")
	posLocation.AttribPointer(2, gl.FLOAT, false, 0, nil)
	posLocation.EnableArray()
	checkGLError()	
	
	imglocation = tonemap.GetUniformLocation("img")
	imglocation.Uniform1i(0)
	checkGLError()	
	
	Flocation := tonemap.GetUniformLocation("F")
	Flocation.Uniform1i(3)
	checkGLError()	
	
	return nil
}
