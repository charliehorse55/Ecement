package Ecement

import (
	"image"
	"os"
	"fmt"
	"image/color"
	_ "image/png"
	"image/jpeg"
    gl "github.com/go-gl/gl"
)

type Pixel struct {
	R, G, B, A float32
}

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

func loadImage(filename string, n *image.NRGBA) error {
	r, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer r.Close()
	
	img, _, err := image.Decode(r)
	if err != nil {
		return err
	}
	
	expectSize := n.Bounds().Max
	imageSize := img.Bounds().Max
	if imageSize.X != expectSize.X || imageSize.Y != expectSize.Y {
		return fmt.Errorf("Image %s has different size", filename)
	}
	
	for j := 0; j < imageSize.Y; j++ {
		for k := 0; k < imageSize.X; k++ {
			n.Set(k, imageSize.Y - (j + 1), img.At(k,j))	
		}
	}
	return nil
}

func loadImages(p Painting, width, height int, removeBackground bool) (error) {

	//create a temporary texture to load images into
	tmpTexture := gl.GenTexture()
	defer tmpTexture.Delete()
	tmpTexture.Bind(gl.TEXTURE_2D)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	
	//create a 1x1 black texture to preprocess the base image with
	//this way we won't subtract anything from the base image
	smallBlackTex := gl.GenTexture()
	defer smallBlackTex.Delete()
	
	//image to subtract from is in texture unit 1
	gl.ActiveTexture(gl.TEXTURE0 + gl.GLenum(1))
	smallBlackTex.Bind(gl.TEXTURE_2D)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	
	pixel := []uint8{0, 0, 0, 255}
    gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, 1, 1, 0, gl.RGBA, gl.UNSIGNED_BYTE, pixel)
			
	pixelBuf := image.NewNRGBA(image.Rectangle{Max: image.Point{X:width, Y:height}})
	for i := range p {
		err := loadImage(p[i].Filename, pixelBuf)
		if err != nil {
			return err
		}
		
		//after we load in the base image, we need to change the image to subract to be it
		//instead of the 1x1 black texture we used to load in the base image
		if i == 1 && removeBackground {
			gl.ActiveTexture(gl.TEXTURE0 + gl.GLenum(1))
			p[0].Texture.Bind(gl.TEXTURE_2D)
		}
		
		//create the actual texture for image (where the preprocessed result goes)
		//don't overwrite the texture in tex unit 1
		gl.ActiveTexture(gl.TEXTURE0 + gl.GLenum(2))
		p[i].Texture = gl.GenTexture()
		p[i].Texture.Bind(gl.TEXTURE_2D)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA32F, width, height, 0, gl.RGBA, gl.FLOAT, nil)
		
		//bind it as the framebuffer target
	 	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, p[i].Texture, 0)	
		
		//upload the image into the temporary GPU texture
		gl.ActiveTexture(gl.TEXTURE0)
		tmpTexture.Bind(gl.TEXTURE_2D)
	    gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, width, height, 0, gl.RGBA, gl.UNSIGNED_BYTE, pixelBuf.Pix)
				
		//preprocess the image
		gl.DrawElements(gl.TRIANGLES, 6, gl.UNSIGNED_INT, nil)		
	} 
	
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
