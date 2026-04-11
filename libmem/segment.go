package libmem

/*
#include "bridge.h"
*/
import "C"

import (
	"errors"
	"runtime/cgo"
	"unsafe"
)

// EnumSegments returns a list of all memory segments in the current process.
func EnumSegments() ([]Segment, error) {
	var result []Segment
	h := cgo.NewHandle(&result)
	defer h.Delete()
	hVal := C.uintptr_t(h)
	ret := C.bridge_enum_segments(unsafe.Pointer(&hVal))
	if ret == C.LM_FALSE {
		return nil, errors.New("libmem: failed to enumerate segments")
	}
	return result, nil
}

// EnumSegmentsEx returns a list of all memory segments in the given process.
func EnumSegmentsEx(process *Process) ([]Segment, error) {
	cp := processToC(process)
	var result []Segment
	h := cgo.NewHandle(&result)
	defer h.Delete()
	hVal := C.uintptr_t(h)
	ret := C.bridge_enum_segments_ex(&cp, unsafe.Pointer(&hVal))
	if ret == C.LM_FALSE {
		return nil, errors.New("libmem: failed to enumerate segments")
	}
	return result, nil
}

// FindSegment finds the memory segment containing the given address.
func FindSegment(address uintptr) (Segment, error) {
	var cs C.lm_segment_t
	ret := C.LM_FindSegment(C.lm_address_t(address), &cs)
	if ret == C.LM_FALSE {
		return Segment{}, errors.New("libmem: segment not found")
	}
	return segmentFromC(&cs), nil
}

// FindSegmentEx finds the memory segment containing the given address in the given process.
func FindSegmentEx(process *Process, address uintptr) (Segment, error) {
	cp := processToC(process)
	var cs C.lm_segment_t
	ret := C.LM_FindSegmentEx(&cp, C.lm_address_t(address), &cs)
	if ret == C.LM_FALSE {
		return Segment{}, errors.New("libmem: segment not found")
	}
	return segmentFromC(&cs), nil
}
