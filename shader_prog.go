package libcement

import (
	"os"
	"io/ioutil"
	"fmt"
    "path/filepath"
    gl "github.com/go-gl/gl"
)

var loadedShaders map[string]gl.Shader
var basePath string

func ShaderInit() error {
    dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
    if err != nil {
		return err
    }
	basePath = dir
	loadedShaders = make(map[string]gl.Shader)
	
	return nil
}

func loadShader(filename string, kind gl.GLenum) (gl.Shader, error)  {
	shader, ok := loadedShaders[filename]
	if ok {
		return shader, nil
	}
	
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
	loadedShaders[filename] = newShader
	return newShader, nil
}


func loadProgram(vertexPath string, fragmentPath string) (gl.Program, error) {

	vertexShader, err := loadShader(filepath.Join(basePath, vertexPath), gl.VERTEX_SHADER)
	if err != nil {
		return 0, fmt.Errorf("Failed to load vertex shader: %v", err)
	}
	fragmentShader, err := loadShader(filepath.Join(basePath, fragmentPath), gl.FRAGMENT_SHADER)
	if err != nil {
		return 0, fmt.Errorf("Failed to load fragment shader: %v", err)
	}
	
	newProgram := gl.CreateProgram()
	newProgram.AttachShader(vertexShader)
	newProgram.AttachShader(fragmentShader)
		
	newProgram.Link()
	status := newProgram.Get(gl.LINK_STATUS)
	if status != gl.TRUE {
		newProgram.Delete()
		return 0, fmt.Errorf("Linking failed: %v", newProgram.GetInfoLog())
	}
	
	return newProgram, nil
}