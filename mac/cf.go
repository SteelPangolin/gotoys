package mac

// #include <CoreFoundation/CoreFoundation.h>
// #cgo CFLAGS:  -framework CoreFoundation
// #cgo LDFLAGS: -framework CoreFoundation
import "C"
import "unsafe"

import (
	"time"
)

// Create a UTF-8 Go string from a CFStringRef.
func GoString(s C.CFStringRef) string {
	// fast way (can fail)
	p := C.CFStringGetCStringPtr(s, C.kCFStringEncodingUTF8)
	if p != nil {
		return C.GoString(p)
	}

	// slow way (usually works)
	bufLen := 1 + C.CFStringGetMaximumSizeForEncoding(
		C.CFStringGetLength(s),
		C.kCFStringEncodingUTF8)
	buf := C.malloc(C.size_t(bufLen))
	defer C.free(buf)
	p = (*C.char)(buf)
	if buf == nil {
		panic("buf should not be nil?")
	}
	success := C.CFStringGetCString(s, p, bufLen, C.kCFStringEncodingUTF8)
	if success == C.false {
		panic("CFStringGetCString failed")
	}
	return C.GoString(p)
}

// Create a CFStringRef from a UTF-8 Go string.
func CreateString(s string) C.CFStringRef {
	cstr := C.CString(s)
	defer C.free(unsafe.Pointer(cstr))
	return C.CFStringCreateWithCString(nil, cstr, C.kCFStringEncodingUTF8)
}

var cfEpoch time.Time = time.Date(2001, time.January, 1, 0, 0, 0, 0, time.UTC)

func GoTime(date C.CFDateRef) time.Time {
	secondsSinceEpoch := float64(C.CFDateGetAbsoluteTime(date))
	return cfEpoch.Add(time.Duration(secondsSinceEpoch * float64(time.Second)))
}

func GoBool(b C.CFBooleanRef) bool {
	return C.CFBooleanGetValue(b) == C.true
}

func GoInt64(num C.CFNumberRef) int64 {
	var value int64
	conversionSuccessful := C.CFNumberGetValue(num, C.kCFNumberSInt64Type, unsafe.Pointer(&value))
	if conversionSuccessful == C.false {
		panic("Lossy or out of range number conversion!")
	}
	return value
}

func GoFloat64(num C.CFNumberRef) float64 {
	var value float64
	conversionSuccessful := C.CFNumberGetValue(num, C.kCFNumberFloat64Type, unsafe.Pointer(&value))
	if conversionSuccessful == C.false {
		panic("Lossy or out of range number conversion!")
	}
	return value
}

func GoNum(num C.CFNumberRef) interface{} {
	if C.CFNumberIsFloatType(num) == C.true {
		return GoFloat64(num)
	} else {
		return GoInt64(num)
	}
}

// Assumes CFTypeRef values.
func GoSlice(array C.CFArrayRef) []interface{} {
	count := int(C.CFArrayGetCount(array))
	slice := make([]interface{}, count, count)
	values := make([]unsafe.Pointer, count)
	allElements := C.CFRange{
		location: 0,
		length:   C.CFIndex(count),
	}
	C.CFArrayGetValues(array, allElements, &values[0])
	for i := 0; i < count; i++ {
		cf := C.CFTypeRef(values[i])
		slice[i] = GoInterface(cf)
	}
	return slice
}

// Assumes CFStringRef keys and CFTypeRef values.
func GoMap(dict C.CFDictionaryRef) map[string]interface{} {
	count := C.CFDictionaryGetCount(dict)
	m := make(map[string]interface{}, count)
	keys := make([]unsafe.Pointer, count)
	values := make([]unsafe.Pointer, count)
	C.CFDictionaryGetKeysAndValues(dict, &keys[0], &values[0])
	for i, key := range keys {
		cfKey := C.CFStringRef(key)
		cfValue := C.CFTypeRef(values[i])
		m[GoString(cfKey)] = GoInterface(cfValue)
	}
	return m
}

// We may need to do this for anything with a CF*GetTypeID function.
// see https://developer.apple.com/library/mac/documentation/CoreFoundation/Reference/CoreFoundation_Collection/_index.html#//apple_ref/doc/uid/TP40003849
func GoInterface(cf C.CFTypeRef) interface{} {
	typeId := C.CFGetTypeID(cf)
	switch typeId {
	case C.CFNullGetTypeID():
		return nil
	case C.CFStringGetTypeID():
		return GoString(C.CFStringRef(cf))
	case C.CFDateGetTypeID():
		return GoTime(C.CFDateRef(cf))
	case C.CFBooleanGetTypeID():
		return GoBool(C.CFBooleanRef(cf))
	case C.CFNumberGetTypeID():
		return GoNum(C.CFNumberRef(cf))
	case C.CFArrayGetTypeID():
		return GoSlice(C.CFArrayRef(cf))
	case C.CFDictionaryGetTypeID():
		return GoMap(C.CFDictionaryRef(cf))
	default:
		// use the equivalent of .String() for CF types
		desc := C.CFCopyDescription(cf)
		defer C.CFRelease(C.CFTypeRef(desc))
		return GoString(desc)
	}
}
