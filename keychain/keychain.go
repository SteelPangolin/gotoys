package keychain

// #include <CoreFoundation/CoreFoundation.h>
// #cgo CFLAGS:  -framework CoreFoundation
// #cgo LDFLAGS: -framework CoreFoundation
// #include <Security/Security.h>
// #cgo CFLAGS:  -framework Security
// #cgo LDFLAGS: -framework Security
import "C"
import "errors"
import "fmt"
import "unsafe"

// Create a UTF-8 Go string from a CFStringRef.
func goStr(s C.CFStringRef) string {
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
        return "" // TODO
    }
    success := C.CFStringGetCString(s, p, bufLen, C.kCFStringEncodingUTF8)
    if success == C.false {
        return "" // TODO
    }
    return C.GoString(p)
}

// Convert an OSStatus from a Security function into a Go error.
func ossErr(st C.OSStatus) error {
    // not an error
    if st == C.noErr {
        return nil
    }

    s := C.SecCopyErrorMessageString(st, nil)
    if s == nil {
        return fmt.Errorf("Unknown error: OSStatus %#v", st)
    }
    defer C.CFRelease(C.CFTypeRef(s))
    return errors.New(goStr(s))
}

func GetVersion() (uint32, error) {
    var version C.UInt32
    st := C.SecKeychainGetVersion(&version)
    return uint32(version), ossErr(st)
}

func Open(path string) (C.SecKeychainRef, error) {
    var kc C.SecKeychainRef
    cpath := C.CString(path)
    defer C.free(unsafe.Pointer(cpath))
    st := C.SecKeychainOpen(cpath, &kc)
    err := ossErr(st)
    return kc, err
}
