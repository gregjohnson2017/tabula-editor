package gfx

import (
	"github.com/go-gl/gl/v2.1/gl"
	"github.com/gregjohnson2017/tabula-editor/pkg/log"
)

type FrameBuffer struct {
	id  uint32
	tex Texture
}

const ErrFrameBuffer log.ConstErr = "incomplete framebuffer"

// NewFrameBuffer creates an FBO of the specified size that renders to
// a texture
func NewFrameBuffer(width, height int32) (FrameBuffer, error) {
	var fb FrameBuffer
	var err error
	gl.GenFramebuffers(1, &fb.id)
	fb.Bind()
	bufs := uint32(gl.COLOR_ATTACHMENT0)
	gl.DrawBuffers(1, &bufs)

	fb.tex, err = NewTexture(width, height, nil, gl.RGBA, 4, 4)
	if err != nil {
		return FrameBuffer{}, err
	}
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, fb.tex.id, 0)

	status := gl.CheckFramebufferStatus(gl.FRAMEBUFFER)
	fb.Unbind()
	if status != gl.FRAMEBUFFER_COMPLETE {
		return FrameBuffer{}, ErrFrameBuffer
	}
	return fb, nil
}

func (fb FrameBuffer) GetTexture() Texture {
	return fb.tex
}

// Bind tells OpenGL to use this framebuffer instead of the default one
func (fb FrameBuffer) Bind() {
	gl.BindFramebuffer(gl.FRAMEBUFFER, fb.id)
}

// Unbind tells OpenGL to stop rendering to this FBO and go back to the screen
func (fb FrameBuffer) Unbind() {
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
}
