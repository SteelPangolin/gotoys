package mac

// #include <CoreFoundation/CoreFoundation.h>
// #cgo CFLAGS:  -framework CoreFoundation
// #cgo LDFLAGS: -framework CoreFoundation
import "C"
import "unsafe"

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

// Create a Go map from a CFDictionaryRef.
// Assumes CFStringRef keys and CFTypeRef values.
func GoMap(dict C.CFDictionaryRef) map[string]string {
    count := C.CFDictionaryGetCount(dict)
    m := make(map[string]string, count)
    keys := make([]unsafe.Pointer, count)
    values := make([]unsafe.Pointer, count)
    C.CFDictionaryGetKeysAndValues(dict, &keys[0], &values[0])
    for i, key := range keys {
        m[GoString(C.CFStringRef(key))] = GoString(C.CFCopyDescription(C.CFTypeRef(values[i])))
    }
    return m
}
