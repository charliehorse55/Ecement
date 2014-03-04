package main

import (
	"image"
	"os"
	"fmt"
	_ "image/png"
	_ "image/jpeg"
    gl "github.com/go-gl/gl"
)

func loadImages(filenames []string) error {
	//setup
	file, err := os.Open(filenames[0])
	if err != nil {
		return err
	}
	config, _, err := image.DecodeConfig(file)
	if err != nil {
		file.Close()
		return err
	}
	file.Close()
	finalBounds := image.Rectangle{Max: image.Point{X:config.Width*len(filenames), Y:config.Height}}
	output := image.NewRGBA(finalBounds)
	
	for i, filename := range filenames {
		r, err := os.Open(filename)
		if err != nil {
			return err
		}
		defer file.Close()
		
		img, _, err := image.Decode(r)
		if err != nil {
			return err
		}
		
		bounds := img.Bounds()
		width := bounds.Max.X
		height := bounds.Max.Y
		if width != config.Width || height != config.Height {
			return fmt.Errorf("Image %s has different size", filename)
		}
		for j := 0; j < height; j++ {
			for k := 0; k < width; k++ {
				output.Set(k + width*i, height - (j + 1), img.At(k,j))	
			}
		}
	} 
		
	texture := gl.GenTexture()
	texture.Bind(gl.TEXTURE_2D)
    gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
    gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, finalBounds.Max.X, finalBounds.Max.Y, 0, gl.RGBA, gl.UNSIGNED_BYTE, output.Pix)
	
	return nil
}		
