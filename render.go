package main

import (
    gl "github.com/go-gl/gl"
)

type OffscreenRender struct {
	fb gl.Framebuffer
	texture gl.Texture
}

func createFramebufferPair(width, height int) []OffscreenRender {
	result := make([]OffscreenRender, 2)
	for i := range result {
		result[i].fb = gl.GenFramebuffer()
		result[i].fb.Bind()
		result[i].texture = gl.GenTexture()
		result[i].texture.Bind(gl.TEXTURE_2D)
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA32F, width, height, 0, gl.RGBA, gl.FLOAT, nil)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	 	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, result[i].texture, 0)	
	}
	return result
}

func render(vectors []lightvector, framebuffers []OffscreenRender, intensity gl.UniformLocation) gl.Texture {
	
	//clear the starting buffer
	framebuffers[1].fb.Bind()
	
	//clear the first source vector
	gl.ClearColor(0.0, 0.0, 0.0, 1.0)
    gl.Clear(gl.COLOR_BUFFER_BIT)
	
	for i,vector := range vectors {
		//update the uniform to use the new intensities
		intensity.Uniform3f(vector.R, vector.G, vector.B)
		
		//bind the previous round framebuffer to texture 0
		gl.ActiveTexture(gl.TEXTURE0)
		framebuffers[(i+1) % 2].texture.Bind(gl.TEXTURE_2D)
		
		//bind the next input image to texture 1
		gl.ActiveTexture(gl.TEXTURE0 + gl.GLenum(1))
		vector.texture.Bind(gl.TEXTURE_2D)
		
		//draw to the output buffer
		framebuffers[i % 2].fb.Bind()
		gl.DrawElements(gl.TRIANGLES, 6, gl.UNSIGNED_INT, nil)
	}
	return framebuffers[(len(vectors)-1) % 2].texture
}
