#version 330

in vec2 texCoord;

layout(location = 0) out vec4 outColor;

uniform sampler2D total;
uniform sampler2D img;

uniform vec3 intensity;

void main() {
	vec4 p = texture(img, texCoord);
	vec4 q = texture(total, texCoord);
	outColor = p*vec4(intensity, 1.0) + q;
//	outColor.a = 1.0;
}

