package gfx

const (
	OutlineVsh = `
	#version 330
	uniform vec4 uni_color;
	uniform vec2 origDims;
	uniform float mult;
	in vec2 position_in;
	out vec4 color;
	void main() {
		vec2 canvasArea = mult * origDims;
		vec2 pos = vec2(mult * position_in.x, canvasArea.y - mult * position_in.y);
		vec2 glSpace = vec2(2.0, 2.0) * (pos / canvasArea) + vec2(-1.0, -1.0);
		gl_Position = vec4(glSpace, 0.0, 1.0);
		color = uni_color;
	}
` + "\x00"

	OutlineFsh = `
	#version 330
	in vec4 color;
	out vec4 frag_color;
	void main() {
		float scale = 4.0;
		float mx = floor(mod(gl_FragCoord.x / scale, 2.0));
		float my = floor(mod(gl_FragCoord.y / scale, 2.0));
		vec4 col1 = vec4(1.0, 1.0, 1.0, 1.0);
		vec4 col2 = vec4(0.3, 0.3, 0.3, 1.0);
		vec4 checker = mx == my ? col1 : col2;
		frag_color = checker;
	}
` + "\x00"

	SolidColorVertex = `
	#version 330
	uniform vec4 uni_color;
	in vec2 position_in;
	out vec4 color;
	void main() {
		gl_Position = vec4(position_in, 0.0, 1.0);
		color = uni_color;
	}
` + "\x00"

	SolidColorFragment = `
	#version 330
	in vec4 color;
	out vec4 frag_color;
	void main() {
		frag_color = color;
	}
` + "\x00"

	VertexShaderSource = `
	#version 330
	uniform vec2 area;
	layout(location = 0) in vec2 position_in;
	layout(location = 1) in vec2 tex_coords_in;
	out vec2 tex_coords;
	void main() {
		vec2 glSpace = vec2(2.0, -2.0) * (position_in / area) + vec2(-1.0, 1.0);
		gl_Position = vec4(glSpace, 0.0, 1.0);
		tex_coords = tex_coords_in;
	}
` + "\x00"

	FragmentShaderSource = `
	#version 330
	uniform sampler2D frag_tex;
	in vec2 tex_coords;
	out vec4 frag_color;
	void main() {
		frag_color = texture(frag_tex, tex_coords);
	}
` + "\x00"

	VshTexturePassthrough = `
	#version 330
	layout(location = 0) in vec2 position_in;
	layout(location = 1) in vec2 tex_coords_in;
	out vec2 tex_coords;
	void main() {
		gl_Position = vec4(position_in, 0.0, 1.0);
		tex_coords = tex_coords_in;
	}
` + "\x00"

	// Uniform `tex_size` is the (width, height) of the texture.
	// Input `position_in` is typical openGL position coordinates.
	// Input `tex_pixels` is the (x, y) of the vertex in the texture starting
	// at (left, top).
	// Output `tex_coords` is typical texture coordinates for fragment shader.
	GlyphShaderVertex = `
	#version 330
	uniform vec2 tex_size;
	uniform vec2 screen_size;
	layout(location = 0) in vec2 position_in;
	layout(location = 1) in vec2 tex_pixels;
	out vec2 tex_coords;
	void main() {
		vec2 glSpace = vec2(2.0, 2.0) * (position_in / screen_size) + vec2(-1.0, -1.0);
		gl_Position = vec4(glSpace, 0.0, 1.0);
		tex_coords = vec2(tex_pixels.x / tex_size.x, tex_pixels.y / tex_size.y);
	}
` + "\x00"

	GlyphShaderFragment = `
	#version 330
	uniform sampler2D frag_tex;
	uniform vec4 text_color;
	in vec2 tex_coords;
	out vec4 frag_color;
	void main() {
		frag_color = vec4(text_color.xyz, texture(frag_tex, tex_coords).r * text_color.w);
	}
` + "\x00"

	CheckerShaderFragment = `
	#version 330
	uniform sampler2D frag_tex;
	in vec2 tex_coords;
	layout(location = 0) out vec4 frag_color;
	void main() {
		float scale = 10.0;
		float mx = floor(mod(gl_FragCoord.x / scale, 2.0));
		float my = floor(mod(gl_FragCoord.y / scale, 2.0));
		vec4 col1 = vec4(1.0, 1.0, 1.0, 1.0);
		vec4 col2 = vec4(0.7, 0.7, 0.7, 1.0);
		vec4 checker = mx == my ? col1 : col2;
		vec4 tex = texture(frag_tex, tex_coords);
		frag_color = mix(checker, tex, tex.a);
	}
` + "\x00"
)
