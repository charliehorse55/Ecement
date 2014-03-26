#version 330

in vec2 texCoord;

layout(location = 0) out vec4 outColor;

uniform sampler2D img;

//lut containing the camera transfer function, F
uniform sampler1D F;

void main() {
	float scale = (256 - 1.0) / 256;
	float offset = 1.0 / (2.0 * 256);
	vec4 p = texture(img, texCoord);
	vec4 data = vec4(sqrt(p.r), sqrt(p.g), sqrt(p.b), 1.0);
	data = clamp(data, vec4(0.0, 0.0, 0.0, 1.0), vec4(1.0, 1.0, 1.0, 1.0));
	data.r = texture(F, (data.r * scale) + offset).r;
	data.g = texture(F, (data.g * scale) + offset).r;
	data.b = texture(F, (data.b * scale) + offset).r;
	outColor = data;
}

