#version 330

in vec2 texCoord;

layout(location = 0) out vec4 outColor;

uniform sampler2D img;

void main() {
	vec4 p = texture(img, texCoord);
	outColor.r = sqrt(p.r);
	outColor.g = sqrt(p.g);
	outColor.b = sqrt(p.b);
	outColor.a = 1.0;
}

