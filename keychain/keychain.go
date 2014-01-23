package keychain

// #cgo CFLAGS:  -framework CoreFoundation -framework Security
// #cgo LDFLAGS: -framework CoreFoundation -framework Security
// #include <CoreFoundation/CoreFoundation.h>
// #include <Security/Security.h>
import "C"
import "errors"

func ossErr(st C.OSStatus) error {
    if st == C.noErr {
        return nil
    }
    return errors.New("some kind of error")
}

func GetVersion() (uint32, error) {
    var version C.UInt32
    st := C.SecKeychainGetVersion(&version)
    return uint32(version), ossErr(st)
}

