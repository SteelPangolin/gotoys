// +build darwin
package mac

// #include <CoreFoundation/CoreFoundation.h>
// #cgo CFLAGS:  -framework CoreFoundation
// #cgo LDFLAGS: -framework CoreFoundation
// #include <CoreServices/CoreServices.h>
// #cgo CFLAGS:  -framework CoreServices
// #cgo LDFLAGS: -framework CoreServices
import "C"

import (
	"fmt"
)

func Spotlight(path string) (map[string]interface{}, error) {
	cfpath := CreateString(path)
	defer C.CFRelease(C.CFTypeRef(cfpath))

	mditem := C.MDItemCreate(nil, cfpath)
	if mditem == nil {
		return nil, fmt.Errorf("MDItemCreate failed for path %#v", path)
	}
	defer C.CFRelease(C.CFTypeRef(mditem))

	names := C.MDItemCopyAttributeNames(mditem)
	if names == nil {
		return nil, fmt.Errorf("MDItemCopyAttributeNames failed for path %#v", path)
	}
	defer C.CFRelease(C.CFTypeRef(names))

	attrs := C.MDItemCopyAttributes(mditem, names)
	if attrs == nil {
		return nil, fmt.Errorf("MDItemCopyAttributes failed for path %#v", path)
	}
	defer C.CFRelease(C.CFTypeRef(attrs))

	return GoMap(attrs), nil
}
