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
	}`

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
	}`

	SolidColorVertex = `
	#version 330
	uniform vec4 uni_color;
	in vec2 position_in;
	out vec4 color;
	void main() {
		gl_Position = vec4(position_in, 0.0, 1.0);
		color = uni_color;
	}`

	SolidColorFragment = `
	#version 330
	in vec4 color;
	out vec4 frag_color;
	void main() {
		frag_color = color;
	}`

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
	}`

	FragmentShaderSource = `
	#version 330
	uniform sampler2D frag_tex;
	in vec2 tex_coords;
	out vec4 frag_color;
	void main() {
		frag_color = texture(frag_tex, tex_coords);
	}`

	VshTexturePassthrough = `
	#version 330
	layout(location = 0) in vec2 position_in;
	layout(location = 1) in vec2 tex_coords_in;
	out vec2 tex_coords;
	void main() {
		gl_Position = vec4(position_in, 0.0, 1.0);
		tex_coords = tex_coords_in;
	}`

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
	}`

	GlyphShaderFragment = `
	#version 330
	uniform sampler2D frag_tex;
	uniform vec4 text_color;
	in vec2 tex_coords;
	out vec4 frag_color;
	void main() {
		frag_color = vec4(text_color.xyz, texture(frag_tex, tex_coords).r * text_color.w);
	}`

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
	}`

	ComputeSelsPerChunk = `
	#version 430
	layout(local_size_x = 1, local_size_y = 1) in;
	layout(std430, binding = 0) buffer ssbo_data
	{
		int chunkIndices[];
	};
	uniform sampler2D selTex;

	void main() {
		ivec2 dims = textureSize(tex, 0);
		ivec2 coords = ivec2(gl_GlobalInvocationID.xy);
		int max_x = max(coords.x * 10 + 10, dims.x);
		int max_y = max(coords.y * 10 + 10, dims.y);
		int n_selected = 0;
		for (int x = coords.x * 10; x < max_x; x++) {
			for (int y = coords.y * 10; y < max_y; y++) {
				if (texelFetch(selTex, ivec2(x, y), 0).x == 1) {
					n_selected++;
				}
			}
		}
		chunkIndices[nXWorkers * coords.y + coords.x] = n_selected;
	}`

	ComputeIndices = `
	#version 430
	layout(local_size_x = 1, local_size_y = 1) in;
	layout(std430, binding = 0) buffer ssbo_data
	{
		int chunkIndices[];
	};

	void main() {
		int currentIdx = 0;
		for (int i = 0; i < chunkIndices.length() - 1; i++) {
			int temp = chunkIndices[i];
			chunkIndices[i] = currentIdx;
			currentIdx += temp;
		}
	}`

	VshPassthrough = `
	#version 330
	in vec2 position_in;
	void main() {
		gl_Position = vec4(position_in, 0.0, 1.0);
	}`

	OutlineGeometry = `
	#version 330
	layout(points) in;
	layout(line_strip, max_vertices=8) out;
	uniform sampler2D selTex;
	uniform vec2 layerArea;
	uniform vec4 view;

	vec2 toGlSpace(vec2 pos) {
		return vec2(2.0, -2.0) * (pos / view.zw) + vec2(-1.0, 1.0);
	}

	void emitLine(vec2 v1, vec2 v2) {
		gl_Position = vec4(v1, 0.0, 1.0);
		EmitVertex();
		gl_Position = vec4(v2, 0.0, 1.0);
		EmitVertex();
		EndPrimitive();
	}

	void main() {
		vec2 texel = gl_in[0].gl_Position.xy;
		vec2 texelPos = layerArea - view.xy + texel;
		
		vec2 tl = toGlSpace(vec2(texelPos.x, texelPos.y));
		vec2 bl = toGlSpace(vec2(texelPos.x, texelPos.y + 1));
		vec2 tr = toGlSpace(vec2(texelPos.x + 1, texelPos.y));
		vec2 br = toGlSpace(vec2(texelPos.x + 1, texelPos.y + 1));
		// emitLine(tl, bl);
		// emitLine(tr, br);
		// emitLine(tl, tr);
		// emitLine(bl, br);
		// left line
		if (texel.x == 0 || texelFetch(selTex, ivec2(texel.x-1, texel.y), 0).r == 0) {
		// 	if (texelPix.x >= intersect.x && texelPix.x <= intersect.x + intersect.z) {
				emitLine(tl, bl);
		// 	}
		}
		// right line
		if (texel.x == textureSize(selTex, 0).x - 1 || texelFetch(selTex, ivec2(texel.x+1, texel.y), 0).r == 0) {
		// 	if (texelPix.x + mult >= intersect.x && texelPix.x + mult <= intersect.x + intersect.z) {
				emitLine(tr, br);
		// 	}
		}
		// top line
		if (texel.y == 0 || texelFetch(selTex, ivec2(texel.x, texel.y-1), 0).r == 0) {
		// 	if (texelPix.y >= intersect.y && texelPix.y <= intersect.y + intersect.w) {
				emitLine(tl, tr);
		// 	}
		}
		// bottom line
		if (texel.y == textureSize(selTex, 0).y - 1 || texelFetch(selTex, ivec2(texel.x, texel.y+1), 0).r == 0) {
		// 	if (texelPix.y >= intersect.y && texelPix.y <= intersect.y + intersect.w) {
				emitLine(bl, br);
		// 	}
		}
	}

	`
)
