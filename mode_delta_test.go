package main

import (
	"bytes"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateDelta(t *testing.T) {

	var basisChunks []string
	chunks, err := readSignature(bytes.NewReader(signature))
	if err != nil {
		panic(err)
	}

	for _, chunk := range chunks {
		basisChunks = append(basisChunks, string(basisFile[chunk.start:chunk.start+chunk.size]))
	}

	if len(basisChunks) != 7 {
		panic("Wrong amount of chunks in basis file!")
	}

	//Verbose = true

	// Chunk sizes
	// 1. 129
	// 2. 72
	// 3. 102
	// 4. 102
	// 5. 84
	// 6. 54
	// 7. 127

	var (
		dataUnmodified = joinChunks(
			basisChunks...,
		)
		dataModifiedSameLen = joinChunks(
			basisChunks[0],
			basisChunks[1],
			basisChunks[2],
			strings.Replace(basisChunks[3], "DEFAULT 0,", "DE0000000,", 1),
			basisChunks[4],
			basisChunks[5],
			basisChunks[6],
		)
		dataDeleteContent = joinChunks(
			basisChunks[0],
			basisChunks[1],
			basisChunks[2],
			strings.Replace(basisChunks[3], "BIGINT", "", 1),
			basisChunks[4],
			basisChunks[5],
			basisChunks[6],
		)
		dataAddContent = joinChunks(
			basisChunks[0],
			basisChunks[1],
			basisChunks[2],
			strings.Replace(basisChunks[3], "B", "LISÄTTY", 1),
			basisChunks[4],
			basisChunks[5],
			basisChunks[6],
		)
		dataAddedBetweenChunks = joinChunks(
			basisChunks[0],
			basisChunks[1],
			basisChunks[2],
			"Added content",
			basisChunks[3],
			basisChunks[4],
			basisChunks[5],
			basisChunks[6],
		)
		dataChunkRemoved = joinChunks(
			basisChunks[0],
			basisChunks[1],
			basisChunks[2],
			basisChunks[4],
			basisChunks[5],
			basisChunks[6],
		)
		dataAddedToBeginning = joinChunks(
			"Added content",
			basisChunks[0],
			basisChunks[1],
			basisChunks[2],
			basisChunks[3],
			basisChunks[4],
			basisChunks[5],
			basisChunks[6],
		)
		dataAddedToEnd = joinChunks(
			basisChunks[0],
			basisChunks[1],
			basisChunks[2],
			basisChunks[3],
			basisChunks[4],
			basisChunks[5],
			basisChunks[6],
			"Added content",
		)
		dataChunksInDifferentPlaces = joinChunks(
			basisChunks[0],
			basisChunks[1],
			basisChunks[2],
			basisChunks[5],
			basisChunks[6],
			basisChunks[3],
			basisChunks[4],
		)
	)

	var tests = []struct {
		name          string
		signature     []byte
		modified      []byte
		expectedDelta []deltaCommand
	}{
		{
			name:      "Unmodified",
			signature: signature,
			modified:  dataUnmodified,
			expectedDelta: []deltaCommand{
				{
					command: COMMAND_COPY,
					start:   0,
					length:  uint64(len(dataUnmodified)),
				},
			},
		},
		// 1. Modified chunk (keep size equal)
		{
			name:      "Modified chunk same size. Modification in end of chunk.",
			signature: signature,
			modified:  dataModifiedSameLen,
			expectedDelta: []deltaCommand{
				{
					command: COMMAND_COPY,
					start:   0,
					length:  303,
				},
				{
					command: COMMAND_LITERAL,
				},
				{
					command: COMMAND_LITERAL,
				},
				{
					command: COMMAND_COPY,
					start:   489,
					length:  181,
				},
			},
		},
		// 2. Data removed from chunk
		{
			name:      "Data removed from chunk",
			signature: signature,
			modified:  dataDeleteContent,
			expectedDelta: []deltaCommand{
				{
					command: COMMAND_COPY,
					start:   0,
					length:  303,
				},
				{
					command: COMMAND_LITERAL,
				},
				{
					command: COMMAND_COPY,
					start:   405,
					length:  265,
				},
			},
		},
		// 3. Data added to chunk
		{
			name:      "Data added to chunk",
			signature: signature,
			modified:  dataAddContent,
			expectedDelta: []deltaCommand{
				{
					command: COMMAND_COPY,
					start:   0,
					length:  303,
				},
				{
					command: COMMAND_LITERAL,
				},
				{
					command: COMMAND_COPY,
					start:   405,
					length:  265,
				},
			},
		},
		// 4. Data added between two chunks
		{
			name:      "Data added between chunks",
			signature: signature,
			modified:  dataAddedBetweenChunks,
			expectedDelta: []deltaCommand{
				{
					command: COMMAND_COPY,
					start:   0,
					length:  303,
				},
				{
					command: COMMAND_LITERAL,
				},
				{
					command: COMMAND_COPY,
					start:   405,
					length:  265,
				},
			},
		},
		// 5. Whole chunk removed
		{
			name:      "1 chunk removed",
			signature: signature,
			modified:  dataChunkRemoved,
			expectedDelta: []deltaCommand{
				{
					command: COMMAND_COPY,
					start:   0,
					length:  303,
				},
				{
					command: COMMAND_COPY,
					start:   405,
					length:  265,
				},
			},
		},
		// 6. Data added to beginning of a file
		{
			name:      "Data added to beginning of a file",
			signature: signature,
			modified:  dataAddedToBeginning,
			expectedDelta: []deltaCommand{
				{
					command: COMMAND_LITERAL,
				},
				{
					command: COMMAND_COPY,
					start:   129,
					length:  541,
				},
			},
		},
		// 7. Data added to end of a file
		{
			name:      "Data added to end of a file",
			signature: signature,
			modified:  dataAddedToEnd,
			expectedDelta: []deltaCommand{
				{
					command: COMMAND_COPY,
					start:   0,
					length:  543,
				},
				{
					command: COMMAND_LITERAL,
				},
			},
		},
		// 8. Change places between chunks
		{
			name:      "Chunks changed places",
			signature: signature,
			modified:  dataChunksInDifferentPlaces,
			expectedDelta: []deltaCommand{
				{
					command: COMMAND_COPY,
					start:   0,
					length:  303,
				},
				{
					command: COMMAND_COPY,
					start:   489,
					length:  54,
				},
				{
					command: COMMAND_LITERAL,
				},
				{
					command: COMMAND_COPY,
					start:   405,
					length:  84,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var deltaB = new(mockDeltaBuffer)

			deltaBufferConstructor = func() DeltaBuffer {
				return deltaB
			}

			_, err := createDelta(
				bytes.NewReader(tt.signature),
				bytes.NewReader(tt.modified),
			)

			assert.NoError(t, err, "createDelta should not return error")

			if len(deltaB.commands) != len(tt.expectedDelta) {
				assert.FailNow(t, "created deltaCommands should have equal amount than expected", "%d != %d", len(deltaB.commands), len(tt.expectedDelta))
			}

			for i := 0; i < len(tt.expectedDelta); i++ {
				assert.Equal(t, tt.expectedDelta[i].command, deltaB.commands[i].command, "Command should be as expected")
				assert.Equal(t, tt.expectedDelta[i].start, deltaB.commands[i].start, "Copy start should be as expected")
				assert.Equal(t, tt.expectedDelta[i].length, deltaB.commands[i].length, "Copy length should be as expected")
			}
		})
	}
}

const (
	COMMAND_COPY    = "copy"
	COMMAND_LITERAL = "literal"
)

type deltaCommand struct {
	command string

	data []byte

	start, length uint64
}

type mockDeltaBuffer struct {
	commands []deltaCommand

	openCopy      bool
	start, length uint64
}

func (dw *mockDeltaBuffer) Bytes() []byte {
	if dw.openCopy {
		dw.endCopy()
	}

	return nil
}

func (dw *mockDeltaBuffer) AddLiteral(data []byte) {
	if dw.openCopy {
		dw.endCopy()
	}

	dw.commands = append(dw.commands, deltaCommand{
		command: COMMAND_LITERAL,
		data:    data,
	})
}

func (dw *mockDeltaBuffer) AddCopy(start, length uint64) {
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

func (dw *mockDeltaBuffer) endCopy() {
	dw.commands = append(dw.commands, deltaCommand{
		command: COMMAND_COPY,
		start:   dw.start,
		length:  dw.length,
	})

	dw.openCopy = false
	dw.start, dw.length = 0, 0
}

func joinChunks(s ...string) []byte {
	return []byte(strings.Join(s, ""))
}

var (
	signature []byte
	basisFile []byte
)

func init() {
	var err error
	signature, err = base64.StdEncoding.DecodeString(signatureFileB64)
	if err != nil {
		panic("Failed to decode basis signature base64 encoding: " + err.Error())
	}

	basisFile, err = base64.StdEncoding.DecodeString(basisFileB64)
	if err != nil {
		panic("Failed to decode basis file base64 encoding: " + err.Error())
	}
}

const (
	signatureFileB64 = "AAAABwAAAAAAAACBAAAAAC2ry//spDMdirla+/gaC4EtUWhQx24cfwAAAIEAAABIAAAAACxM6X9tk7Q6mgo0hgtuT9OB6ZDJbZnWoAAAAMkAAABmAAAAABKqbn9daNHWNFr/ZOP6X8XhCQuyTC2UggAAAS8AAABmAAAAAANoaH/MEazbZtnrgMYQt3KOSRFviMb38wAAAZUAAABUAAAAAB/BRv/G9252GsKc1S+zExXphAW/lRKk9gAAAekAAAA2AAAAAC2ry//9PsUPI/pvY8zOqocqp0f/EKJXFAAAAh8AAAB/AAAAABGG878Km/PjNJcjJVATS+9NuAeWiHmDKg=="

	basisFileB64 = "c3Rpb25fY2F0ZWdvcnlfc2l0ZV90eXBlcyAoDQogICAgICAgIHF1ZXN0aW9uX2NhdGVnb3J5X2lkIFNNQUxMSU5UIFJFRkVSRU5DRVMgcXVlc3Rpb25fY2F0ZWdvcnkocXVlc3Rpb25fY2F0ZWdvcnlfaWQpIE9OIERFTEVURSBDQVNDQURFLA0KICAgICAgICBzaXRlX3R5cGVfaWQgU01BTExJTlQgUkVGRVJFTkNFUyBzaXRlX3R5cGUoc2l0ZV90eXBlX2lkKSBPTiBERUxFVEUgQ0FTQ0FERSwNCiAgICAgICAgUFJJTUFSWSBLRVkgKHF1ZXN0aW9uX2NhdGVnb3J5X2lkLCBzaXRlX3R5cGVfaWQpDQopOw0KDQpDUkVBVEUgVEFCTEUgcXVlc3Rpb24gKA0KICAgICAgICBxdWVzdGlvbl9pZCBCSUdTRVJJQUwgUFJJTUFSWSBLRVksDQogICAgICAgIHNvcnRfbnVtIEJJR0lOVCBOT1QgTlVMTCBERUZBVUxUIDAsDQogICAgICAgIHF1ZXN0aW9uIFRFWFQgTk9UIE5VTEwsDQogICAgICAgIHF1ZXN0aW9uX2NhdGVnb3J5X2lkIFNNQUxMSU5UIE5PVCBOVUxMIFJFRkVSRU5DRVMgcXVlc3Rpb25fY2F0ZWdvcnkocXVlc3Rpb25fY2F0ZWdvcnlfaWQpIE9OIERFTEVURSBDQVNDQURFLA0KICAgICAgICB3ZWlnaHQgTlVNRVJJQyBOT1QgTlVMTCBERUZBVUxUIDEsDQogICAgICAgIGVuYWJsZWQgQk9PTEVBTiBOT1QgTlVMTCBERUZBVUxUIFRSVUUNCik7DQoNCkNSRUFURSBUQUJMRSBhbnN3ZXJfdA=="
)

/* Contents of basisFileB64 (note that line break characters are CRLF and not LF):
stion_category_site_types (
        question_category_id SMALLINT REFERENCES question_category(question_category_id) ON DELETE CASCADE,
        site_type_id SMALLINT REFERENCES site_type(site_type_id) ON DELETE CASCADE,
        PRIMARY KEY (question_category_id, site_type_id)
);

CREATE TABLE question (
        question_id BIGSERIAL PRIMARY KEY,
        sort_num BIGINT NOT NULL DEFAULT 0,
        question TEXT NOT NULL,
        question_category_id SMALLINT NOT NULL REFERENCES question_category(question_category_id) ON DELETE CASCADE,
        weight NUMERIC NOT NULL DEFAULT 1,
        enabled BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE answer_t
*/

// Just for easier checking of what each chunk contains
/*
const (
	basisChunk1 = `stion_category_site_types (
        question_category_id SMALLINT REFERENCES question_category(question_category_id) ON DELETE C`
	basisChunk2 = `ASCADE,
	site_type_id SMALLINT REFERENCES site_type(site_type_id`
	basisChunk3 = `) ON DELETE CASCADE,
	PRIMARY KEY (question_category_id, site_type_id)
);

CREATE TABLE que`
	basisChunk4 = `stion (
        question_id BIGSERIAL PRIMARY KEY,
        sort_num BIGINT NOT NULL DEFAULT 0,
    `
	basisChunk5 = `    question TEXT NOT NULL,
	question_category_id SMALLINT NOT NULL REFERENC`
	basisChunk6 = `ES question_category(question_category_id) ON DELETE C`
	basisChunk7 = `ASCADE,
	weight NUMERIC NOT NULL DEFAULT 1,
	enabled BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE answer_t`
)
*/
