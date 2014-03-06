package main

import (
	"image"
	"os"
	"fmt"
	"image/color"
	_ "image/png"
	"image/jpeg"
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

func saveToJPEG(filename string, width, height int, data []Pixel) error {
	output := image.NewNRGBA(image.Rectangle{Max: image.Point{X:width, Y:height}})
	for i := 0; i < height; i++ {
		for j := 0; j < width; j++ {
			
			//prevent overflow for each channel when packed into 8 bits
			red := data[i*width +j].R*255
			if red > 255 {
				red = 255
			}
			green := data[i*width +j].G*255
			if green > 255 {
				green = 255
			}
			blue := data[i*width +j].B*255
			if blue > 255 {
				blue = 255
			}
	
			output.Set(j, height - (i+1), color.NRGBA{
					R:uint8(red),
					G:uint8(green),
					B:uint8(blue),
					A:uint8(255),
				})
		}
	}
		
	outfile, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("Failed to open output file: %s", filename)
	}
	defer outfile.Close()

	err = jpeg.Encode(outfile, output, nil)
	if err != nil {
		return fmt.Errorf("Failed to encode output file: %v", err)
	}
	
	return nil
}
