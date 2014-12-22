package dbm

// #include <errno.h>
// #include <ndbm.h>
// #include <stdlib.h>
// #include <string.h>
import "C"

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

type DBM struct {
	cDbm *C.DBM
}

type DBMError struct {
	errno  int
	errstr string
	err    error
}

var AlreadyExists = &DBMError{
	errstr: "Key already exists",
}

var NotFound = &DBMError{
	errstr: "Key not found",
}

func newError(errno C.int) *DBMError {
	err := &DBMError{
		errno: int(errno),
	}
	cStr := C.strerror(errno)
	if uintptr(unsafe.Pointer(cStr)) == C.EINVAL {
		err.errstr = fmt.Sprintf("Unknown error: %d", errno)
	} else {
		err.errstr = C.GoString(cStr)
	}
	return err
}

func wrapError(err error) error {
	return &DBMError{
		err:    err,
		errstr: err.Error(),
	}
}

func (err *DBMError) Error() string {
	return err.errstr
}

func Open(path string) (*DBM, error) {
	cPath := C.CString(path)
	cDbm, err := C.dbm_open(
		cPath,
		C.int(os.O_RDWR|os.O_CREATE),
		syscall.S_IREAD|syscall.S_IWRITE|syscall.S_IRGRP|syscall.S_IWGRP)
	C.free(unsafe.Pointer(cPath))
	if err != nil {
		return nil, wrapError(err)
	}
	dbm := &DBM{
		cDbm: cDbm,
	}
	return dbm, nil
}

func (dbm *DBM) checkError() *DBMError {
	errno := C.dbm_error(dbm.cDbm)
	if errno == 0 {
		return nil
	}
	C.dbm_clearerr(dbm.cDbm)
	return newError(errno)
}

func (dbm *DBM) Close() {
	C.dbm_close(dbm.cDbm)
}

const (
	dbm_SUCCESS        = 0
	dbm_ALREADY_EXISTS = 1
	dbm_NOT_FOUND      = 1
	dbm_ERROR          = -1
)

func bytesToDatum(buf []byte) C.datum {
	return C.datum{
		dptr:  unsafe.Pointer(&buf[0]),
		dsize: C.size_t(len(buf)),
	}
}

func datumToBytes(datum C.datum) []byte {
	if datum.dptr == nil {
		return nil
	}
	return C.GoBytes(datum.dptr, C.int(datum.dsize))
}

func (dbm *DBM) store(key, content []byte, mode C.int) (C.int, error) {
	status, err := C.dbm_store(dbm.cDbm, bytesToDatum(key), bytesToDatum(content), mode)
	if err != nil {
		return status, wrapError(err)
	}
	if status == dbm_ERROR {
		return status, dbm.checkError()
	}
	return status, nil
}

func (dbm *DBM) Insert(key, content []byte) error {
	status, err := dbm.store(key, content, C.DBM_INSERT)
	if err != nil {
		return err
	}
	if status == dbm_ALREADY_EXISTS {
		return nil
	}
	return nil
}

func (dbm *DBM) Replace(key, content []byte) error {
	_, err := dbm.store(key, content, C.DBM_REPLACE)
	return err
}

func (dbm *DBM) Fetch(key []byte) ([]byte, error) {
	datum, err := C.dbm_fetch(dbm.cDbm, bytesToDatum(key))
	if err != nil {
		return nil, wrapError(err)
	}
	return datumToBytes(datum), nil
}

func (dbm *DBM) Delete(key []byte) error {
	status, err := C.dbm_delete(dbm.cDbm, bytesToDatum(key))
	if err != nil {
		println("C error")
		return wrapError(err)
	}
	if status == dbm_ERROR {
		println("DB error")
		return dbm.checkError()
	}
	if status == dbm_NOT_FOUND {
		return NotFound
	}
	return nil
}

func (dbm *DBM) KeysCallback(callback func([]byte) error) error {
	// TODO: rework this loop with C errno support and explicit DB error checks
	for key := datumToBytes(C.dbm_firstkey(dbm.cDbm)); key != nil; key = datumToBytes(C.dbm_nextkey(dbm.cDbm)) {
		err := callback(key)
		if err != nil {
			return err
		}
	}
	return nil
}

func (dbm *DBM) Keys() [][]byte {
	keys := [][]byte{}
	_ = dbm.KeysCallback(func(key []byte) error {
		keys = append(keys, key)
		return nil
	})
	return keys
}

func (dbm *DBM) Len() int {
	count := 0
	_ = dbm.KeysCallback(func(_ []byte) error {
		count++
		return nil
	})
	return count
}

func (dbm *DBM) ValuesCallback(callback func([]byte) error) error {
	return dbm.KeysCallback(func(key []byte) error {
		value, err := dbm.Fetch(key)
		if err != nil {
			return nil
		}
		return callback(value)
	})
}

func (dbm *DBM) Values() [][]byte {
	values := [][]byte{}
	_ = dbm.ValuesCallback(func(value []byte) error {
		values = append(values, value)
		return nil
	})
	return values
}

func (dbm *DBM) ItemsCallback(callback func(key, value []byte) error) error {
	return dbm.KeysCallback(func(key []byte) error {
		value, err := dbm.Fetch(key)
		if err != nil {
			return err
		}
		return callback(key, value)
	})
}

type DBMItem struct {
	Key   []byte
	Value []byte
}

func (dbm *DBM) Items() []DBMItem {
	items := []DBMItem{}
	_ = dbm.ItemsCallback(func(key, value []byte) error {
		items = append(items, DBMItem{
			Key:   key,
			Value: value,
		})
		return nil
	})
	return items
}
