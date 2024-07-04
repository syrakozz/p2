// Package opus interfaces with libopus.a
package opus

/*
#cgo CFLAGS: -I/usr/include/opus
#cgo LDFLAGS: -L/usr/lib -lopus
#include <stdlib.h>
#include <opus.h>

#define CHUNK_SIZE 320
#define MAX_FRAME_SIZE 48000

int go_opus_encoder_ctl(OpusEncoder *opusEncoder)  {
	int err = 0;

    err = opus_encoder_ctl(opusEncoder, OPUS_SET_VBR(0));
    if (err < 0) {
		return err;
    }

    err = opus_encoder_ctl(opusEncoder, OPUS_SET_BITRATE(16000));
    if (err < 0) {
        return err;
    }

    err = opus_encoder_ctl(opusEncoder, OPUS_SET_COMPLEXITY(3));
    if (err < 0) {
        return err;
    }

	return 0;
}

unsigned char* encodeOneFrame(OpusEncoder *opusEncoder, const unsigned char* input, int inputSize, int* encodedLength) {
    unsigned char* encodedFrameData = NULL;

    short* pcm_input = (short*)malloc((inputSize / 2) * sizeof(short));
    unsigned char* encodedData = (unsigned char*)calloc((inputSize / 2), sizeof(unsigned char));

    for (int i = 0; i < (inputSize / 2); i++) {
        opus_int32 s;
        s = input[(2 * i) + 1] << 8 | input[2 * i];
        s = ((s & 0xFFFF) ^ 0x8000) - 0x8000;
        pcm_input[i] = (short)s;
    }

	*encodedLength = opus_encode(opusEncoder, pcm_input, (inputSize / 2), encodedData, (inputSize / 2));

    if (*encodedLength > 0) {
        encodedFrameData = (unsigned char*)malloc(*encodedLength);
        if (encodedFrameData != NULL) {
            for (int i = 0; i < *encodedLength; i++) {
                encodedFrameData[i] = encodedData[i];
            }
        }
    }

    free(pcm_input);
    free(encodedData);

    return encodedFrameData;
}
*/
import "C"

import (
	"errors"
	"fmt"
	"io"
	"unsafe"
)

const (
	// FrameSize is one opus frame.
	FrameSize       = 640
	opusSampleRate  = C.int(16000)
	opusChannels    = C.int(1)
	opusApplication = C.OPUS_APPLICATION_AUDIO
)

// CreateOpusEncoder returns a new encoder.
func CreateOpusEncoder() (*C.OpusEncoder, error) {
	var (
		cErr C.int
	)

	e := C.opus_encoder_create(opusSampleRate, opusChannels, opusApplication, &cErr)
	if int(cErr) != 0 {
		return nil, fmt.Errorf("unable to create opus encoder: %s", C.GoString(C.opus_strerror(cErr)))
	}
	return e, nil
}

// DestroyOpusEncoder frees an encoder.
func DestroyOpusEncoder(opusEncoder *C.OpusEncoder) {
	C.opus_encoder_destroy(opusEncoder)
}

// InitOpusEncoder initializes an encoder.
func InitOpusEncoder(opusEncoder *C.OpusEncoder) error {
	if cErr := C.go_opus_encoder_ctl(opusEncoder); int(cErr) < 0 {
		return fmt.Errorf("unable to set constant bitrate: %s", C.GoString(C.opus_strerror(cErr)))
	}

	return nil
}

// WriteOpusFrame writes one opus frame to the writer.
func WriteOpusFrame(opusEncoder *C.OpusEncoder, w io.Writer, data []byte) error {
	var encodedLength C.int

	cData := C.CBytes(data)
	defer C.free(cData)

	cEncodedData := C.encodeOneFrame(opusEncoder, (*C.uchar)(cData), C.int(len(data)), &encodedLength)
	if int(encodedLength) < 0 {
		return errors.New("unable to encode opus frame")
	}

	goEncodedData := C.GoBytes(unsafe.Pointer(cEncodedData), encodedLength)
	if _, err := w.Write(goEncodedData); err != nil {
		return errors.New("unable to write opus data to stream")
	}

	return nil
}
