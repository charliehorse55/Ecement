package main

import (
	"image"
	"os"
	"fmt"
	_ "image/png"
	_ "image/jpeg"
    gl "github.com/go-gl/gl"
)

func getImageRes(filename string) (int, int, error) {
	file, err := os.Open(filename)
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()
	
	config, _, err := image.DecodeConfig(file)
	if err != nil {
		return 0, 0, err
	}
	
	return config.Width, config.Height, nil
}

func loadImages(filenames []string, Width, Height int) error {
	//setup
	finalBounds := image.Rectangle{Max: image.Point{X:Width*len(filenames), Y:Height}}
	output := image.NewRGBA(finalBounds)
	
	for i, filename := range filenames {
		r, err := os.Open(filename)
		if err != nil {
			return err
		}
		defer r.Close()
		
		img, _, err := image.Decode(r)
		if err != nil {
			return err
		}
		
		bounds := img.Bounds()
		width := bounds.Max.X
		height := bounds.Max.Y
		if width != Width || height != Height {
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
