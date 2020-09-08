package gfx

import (
	"unsafe"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/gregjohnson2017/tabula-editor/pkg/log"
)

type BufferObject struct {
	id        uint32
	sizeBytes uint32
}

func NewBufferObject() *BufferObject {
	var bo BufferObject
	gl.GenBuffers(1, &bo.id)
	bo.sizeBytes = 0
	return &bo
}

func (bo *BufferObject) BufferData(target uint32, sizeBytes uint32, ptr unsafe.Pointer, usage uint32) {
	bo.sizeBytes = sizeBytes
	bo.Bind(target)
	gl.BufferData(target, int(sizeBytes), ptr, usage)
	bo.Unbind(target)
}

func (bo *BufferObject) BufferSubData(target, offset, sizeBytes uint32, ptr unsafe.Pointer) {
	// gl.BufferData acts like malloc, while gl.BufferSubData acts like memcpy
	// BufferSubData can only modify a range of the existing size
	if offset+sizeBytes > bo.sizeBytes {
		log.Warn("BufferSubData out of bounds")
	}
	bo.Bind(target)
	gl.BufferSubData(target, int(offset), int(sizeBytes), ptr)
	bo.Unbind(target)
}

func (bo *BufferObject) GetSubData(target, offset, sizeBytes uint32, ptr unsafe.Pointer) {
	bo.Bind(target)
	gl.GetBufferSubData(target, int(offset), int(sizeBytes), ptr)
	bo.Unbind(target)
}
func (bo *BufferObject) GetData(target uint32, ptr unsafe.Pointer) {
	bo.GetSubData(target, 0, bo.sizeBytes, ptr)
}
func (bo *BufferObject) GetID() uint32 {
	return bo.id
}

func (bo *BufferObject) GetSizeBytes() uint32 {
	return bo.sizeBytes
}

func (bo *BufferObject) Bind(target uint32) {
	gl.BindBuffer(target, bo.id)
}

func (bo *BufferObject) Unbind(target uint32) {
	gl.BindBuffer(target, 0)
}

func (bo *BufferObject) BindBufferBase(target, binding uint32) {
	gl.BindBufferBase(target, binding, bo.id)
}

func (bo *BufferObject) Destroy() {
	gl.DeleteBuffers(1, &bo.id)
	bo.id = 0
	bo.sizeBytes = 0
}
