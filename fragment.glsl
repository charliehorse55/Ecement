#version 330

in vec2 texCoord;

out vec4 outColor;

uniform sampler2D tex;
uniform int num;

uniform float[10] intensity;

void main() {
	outColor = vec4(0.0, 0.0, 0.0, 1.0);
	for(int i = 0; i < num; i++) {
		vec4 p = texture(tex, vec2((texCoord.x/num) + (float(i)/float(num)), texCoord.y));
		outColor.r += (p.r*p.r)*intensity[i];
		outColor.g += (p.g*p.g)*intensity[i];
		outColor.b += (p.b*p.b)*intensity[i];
	}
	outColor.r = sqrt(outColor.r);
	outColor.g = sqrt(outColor.g);
	outColor.b = sqrt(outColor.b);
}

