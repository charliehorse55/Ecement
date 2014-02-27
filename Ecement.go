package main

import (
	"image"
	"image/color"
	"flag"
	"os"
	"log"
	"math"
	_ "image/png"
	"image/jpeg"
)

type Pixel struct {
	R, G, B, A float64
}

func main() {
	flag.Parse()
	
	imagePaths := flag.Args()
	
	if len(imagePaths) < 2 {
		log.Printf("Ecement requires at least 2 input files")
		return
	}	
	
	var outputSize image.Rectangle
	var outputImage []Pixel
	for _,filename := range imagePaths {
		file, err := os.Open(filename)
		if err != nil {
			log.Printf("Failed to open %s: %v", filename, err)
			continue
		}
		defer file.Close()
		
		img, _,err := image.Decode(file)
		if err != nil {
			log.Printf("Image file %s corrupt: %v", filename, err)
		}
		
		bounds := img.Bounds()
		if outputSize.Max.X == 0 {
			outputSize = bounds
			outputImage = make([]Pixel, bounds.Max.X * bounds.Max.Y)
		}

		if outputSize != bounds {
			log.Printf("Image %s does not match size of first image", filename)
			return
		}
		
		//add the sum of squares
		row := outputSize.Max.Y
		col := outputSize.Max.X
		for i := 0; i < row; i++ {
			for j := 0; j < col; j++ {
				r, g, b, _ := img.At(j,i).RGBA()			
				outputImage[i*col + j].R += float64(r*r)
				outputImage[i*col + j].G += float64(g*g)
				outputImage[i*col + j].B += float64(b*b)
			}
		}
	}
		
	//save the image
	output := image.NewRGBA(outputSize)
	row := outputSize.Max.Y
	col := outputSize.Max.X
	for i := 0; i < row; i++ {
		for j := 0; j < col; j++ {
			R := uint32(math.Sqrt(outputImage[i*col + j].R))/256
			G := uint32(math.Sqrt(outputImage[i*col + j].G))/256
			B := uint32(math.Sqrt(outputImage[i*col + j].B))/256
						
			//cap values to 255 to fit into 8 bit color channels
			if R > 255 {
				R = 255
			}
			if G > 255 {
				G = 255
			}
			if B > 255 {
				B = 255
			}
			
			output.Set(j, i, color.RGBA{
				R:uint8(R),
				G:uint8(G),
				B:uint8(B),
				A:uint8(255),
			})
		}
	}
	
	
	outfilepath := "output.jpg"
	outfile, err := os.Create(outfilepath)
	if err != nil {
		log.Printf("Failed to open output file: %s", outfilepath)
		return 
	}
	defer outfile.Close()
	
	err = jpeg.Encode(outfile, output, nil)
	if err != nil {
		log.Printf("Failed to encode output file")
	}
	
}