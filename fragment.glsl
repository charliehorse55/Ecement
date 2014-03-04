#version 330

in vec2 texCoord;

out vec4 outColor;

uniform sampler2D tex0;
uniform sampler2D tex1;

uniform vec2 intensity;

void main() {
	vec4 p0 = texture(tex0, texCoord);
	vec4 p1 = texture(tex1, texCoord);
    outColor.r = sqrt((p0.r*p0.r)*intensity.x + (p1.r*p1.r)*intensity.y);
    outColor.g = sqrt((p0.g*p0.g)*intensity.x + (p1.g*p1.g)*intensity.y);
    outColor.b = sqrt((p0.b*p0.b)*intensity.x + (p1.b*p1.b)*intensity.y);
	outColor.a = 1.0;
}

