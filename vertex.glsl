#version 330

in vec2 position;
out vec2 texCoord;

void main() {
	texCoord = vec2(position.x, position.y);
	
	//convert from 0,0 to 1,1 coordinates to -1,-1 to 1,1 coords
	gl_Position = vec4(position.x*2 - 1.0, position.y*2 - 1.0, 0.0, 1.0);
}
