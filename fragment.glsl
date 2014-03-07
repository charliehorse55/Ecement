#version 330

in vec2 texCoord;

layout(location = 0) out vec4 outColor;

uniform sampler2D old;
uniform sampler2D new;

uniform vec3 intensity;

void main() {
	vec4 p = texture(new, texCoord);
	vec4 q = texture(old, texCoord);
	outColor.r = (p.r*p.r)*intensity.r + q.r;
	outColor.g = (p.g*p.g)*intensity.g + q.g;
	outColor.b = (p.b*p.b)*intensity.b + q.b;
	outColor.a = 1.0;
}

