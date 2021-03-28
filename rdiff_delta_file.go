package main

import (
	"bytes"
	"encoding/binary"
)

const (
	RS_DELTA_MAGIC = "rs\x026"

	// Inform integer sizes as uint64
	RS_OP_LITERAL_N8 = uint8(0x44)

	// Inform integer sizes as uint64
	RS_OP_COPY_N8_N8 = uint8(0x54)
)

// RdiffDelta constructs rdiff delta file to inner buffer
type RdiffDelta struct {
	b *bytes.Buffer

	openCopy bool
	start    uint64
	length   uint64
}

// NewDeltaBuffer initiates delta file buffer
func NewRdiffDelta() DeltaBuffer {
	dw := &RdiffDelta{
		b: new(bytes.Buffer),
	}
	dw.b.Write([]byte(RS_DELTA_MAGIC))

	return dw
}

// Bytes closes the buffer and returns bytes from buffer
func (dw *RdiffDelta) Bytes() []byte {
	if dw.openCopy {
		dw.endCopy()
	}

	// Write end command
	dw.b.WriteByte(0x00)
	return dw.b.Bytes()
}

// AddLiteral writes literal command to buffer
func (dw *RdiffDelta) AddLiteral(data []byte) {
	if dw.openCopy {
		dw.endCopy()
	}

	dw.b.WriteByte(RS_OP_LITERAL_N8)

	binary.Write(dw.b, binary.BigEndian, uint64(len(data)))
	dw.b.Write(data)
}

// AddCopy writes copy command to buffer
func (dw *RdiffDelta) AddCopy(start, length uint64) {
	if dw.openCopy && dw.start+dw.length != start {
		dw.endCopy()
	}

	if !dw.openCopy {
		dw.openCopy = true
		dw.start = start
		dw.length = 0
	}
	dw.length += length
}

// endCopy writes the combined COPY command to buffer
func (dw *RdiffDelta) endCopy() {
	dw.b.WriteByte(RS_OP_COPY_N8_N8)

	binary.Write(dw.b, binary.BigEndian, dw.start)
	binary.Write(dw.b, binary.BigEndian, dw.length)

	dw.openCopy = false
	dw.start, dw.length = 0, 0
}
