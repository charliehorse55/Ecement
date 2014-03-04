package main

import (
	"image"
	"io"
	"fmt"
	_ "image/png"
	_ "image/jpeg"
    gl "github.com/go-gl/gl"
)

func loadImage(r io.Reader, index int) (gl.Texture, error) {
	img, _, err := image.Decode(r)
	if err != nil {
		return 0, fmt.Errorf("Image file corrupt: %v", err)
	}
	bounds := img.Bounds()
	output := image.NewRGBA(bounds)
	height := bounds.Max.Y
	width := bounds.Max.X
	for i := 0; i < height; i++ {
		for j := 0; j < width; j++ {
			output.Set(j, i, img.At(j,i))	
		}
	}
	
	gl.ActiveTexture(gl.TEXTURE0 + gl.GLenum(index))
	texture := gl.GenTexture()
	texture.Bind(gl.TEXTURE_2D)
    gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
    gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, width, height, 0, gl.RGBA, gl.UNSIGNED_BYTE, output.Pix)
	
	return texture, nil
}		
