package gfx

const (
	OutlineVsh = `
	#version 330
	uniform vec2 area;
	layout(location = 0) in vec2 position_in;
	void main() {
		vec2 glSpace = vec2(2.0, -2.0) * (position_in / area) + vec2(-1.0, 1.0);
		gl_Position = vec4(glSpace, 0.0, 1.0);
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

	SolidColorFragment = `
	#version 330
	uniform vec4 uni_color;
	out vec4 frag_color;
	void main() {
		frag_color = uni_color;
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
		
		// left line
		if (texel.x == 0 || texelFetch(selTex, ivec2(texel.x-1, texel.y), 0).r == 0) {
			if (texelPos.x >= 0 && texelPos.x <= view.z) {
				emitLine(tl, bl);
			}
		}
		// right line
		if (texel.x == textureSize(selTex, 0).x - 1 || texelFetch(selTex, ivec2(texel.x+1, texel.y), 0).r == 0) {
			if (texelPos.x + 1 >= 0 && texelPos.x + 1 <= view.z) {
				emitLine(tr, br);
			}
		}
		// top line
		if (texel.y == 0 || texelFetch(selTex, ivec2(texel.x, texel.y-1), 0).r == 0) {
			if (texelPos.y >= 0 && texelPos.y <= view.w) {
				emitLine(tl, tr);
			}
		}
		// bottom line
		if (texel.y == textureSize(selTex, 0).y - 1 || texelFetch(selTex, ivec2(texel.x, texel.y+1), 0).r == 0) {
			if (texelPos.y + 1 >= 0 && texelPos.y + 1 <= view.w) {
				emitLine(bl, br);
			}
		}
	}`

	ComputeCountSels = `
	#version 430
	layout(local_size_x = 1) in;
	layout(std430, binding = 0) buffer setup
	{
		uint chunkIndices[];
	};
	layout(std430, binding = 2) buffer offsets
	{
		uint sum;
		uint next[];
	};
	uniform usampler2D selTex;
	uniform uint chunkSize;

	void main() {
		uvec2 dims = uvec2(textureSize(selTex, 0));
		uint wid = uint(gl_WorkGroupID.x);
		uint edge = 0;
		if (wid == gl_NumWorkGroups.x - 1) {
			edge = dims.x*dims.y - chunkSize*gl_NumWorkGroups.x;
		}
		for (uint i = wid * chunkSize; i < (wid+1)*chunkSize+edge; i++) {
			if (i == wid * chunkSize) {
				chunkIndices[wid] = 0;
				next[wid] = 0;
			}
			uint tx = i % dims.x;
			uint ty = i / dims.x;
			if (texelFetch(selTex, ivec2(tx,ty), 0).r != 0) {
				chunkIndices[wid] += 1;
				next[wid] += 1;
			}
		}
	}`

	ComputePrefixSum = `
	#version 430
	layout(local_size_x = 1) in;
	layout(std430, binding = 0) buffer setup
	{
		uint chunkOffs[];
	};
	layout(std430, binding = 2) buffer offsets
	{
		uint sum;
		uint next[];
	};
	uniform int pass;

	void main() {
		uint wid = uint(gl_WorkGroupID.x);
		if (pass == uint(ceil(log2(gl_NumWorkGroups.x)))) {
			if (pass % 2 == 0) { 
				if (wid == gl_NumWorkGroups.x - 1) {
					sum = chunkOffs[wid]*2;
				}
				next[wid] = wid == 0 ? 0 : (chunkOffs[wid-1]) * 2;
			} else {
				if (wid == gl_NumWorkGroups.x - 1) {
					sum = next[wid]*2;
				}
				chunkOffs[wid] = wid == 0 ? 0 : (next[wid-1]) * 2;
			}
		} else {
			if (pass % 2 == 0) {
				if (wid < uint(exp2(pass))) {
					next[wid] = chunkOffs[wid];
				} else {
					next[wid] = chunkOffs[wid] + chunkOffs[wid-uint(exp2(pass))];
				}
			} else {
				if (wid < uint(exp2(pass))) {
					chunkOffs[wid] = next[wid];
				} else {
					chunkOffs[wid] = next[wid] + next[wid-uint(exp2(pass))];
				}
			}
		}
	}`

	ComputeSelCoords = `
	#version 430
	layout(local_size_x = 1) in;
	layout(std430, binding = 2) buffer offsets
	{
		uint sum;
		uint next[];
	};
	layout(std430, binding = 1) buffer verts
	{
		float finalAnswer[];
	};
	layout(std430, binding = 0) buffer setup
	{
		uint chunkOffs[];
	};
	uniform usampler2D selTex;
	uniform uint chunkSize;

	void main() {
		uvec2 dims = uvec2(textureSize(selTex, 0));
		uint wid = uint(gl_WorkGroupID.x);
		uint edge = 0;
		if (wid == gl_NumWorkGroups.x - 1) {
			edge = dims.x*dims.y - chunkSize*gl_NumWorkGroups.x;
		}
		uint passes = uint(ceil(log2(chunkOffs.length())));
		uint off;
		if (passes%2==0) {
			off = next[wid];
		} else {
			off = chunkOffs[wid];
		}
		uint j = 0;
		for (uint i = wid * chunkSize; i < (wid+1)*chunkSize+edge; i++) {
			uint tx = i % dims.x;
			uint ty = i / dims.x;
			if (texelFetch(selTex, ivec2(tx,ty), 0).r != 0) {
				finalAnswer[off+j] = float(tx);
				finalAnswer[off+j+1] = float(ty);
				j += 2;
			}
		}
	}`
)
