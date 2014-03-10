#version 330

in vec2 texCoord;

layout(location = 0) out vec4 outColor;

uniform sampler2D img;
uniform sampler2D base;

void main() {
	vec4 p = texture(img, texCoord);
	vec4 q = texture(base, texCoord);
	outColor = p*p - q;
	outColor = max(outColor, vec4(0.0, 0.0, 0.0, 1.0));
}

