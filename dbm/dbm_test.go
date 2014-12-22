package dbm

import (
	"bytes"
	"io/ioutil"
	"os"
	//"sort"
	"path/filepath"
	"testing"
)

type DBMItems []DBMItem

func (items DBMItems) Len() int {
	return len(items)
}

func (items DBMItems) Swap(i, j int) {
	items[i], items[j] = items[j], items[i]
}

func (items DBMItems) Less(i, j int) bool {
	return bytes.Compare(items[i].Key, items[j].Key) == -1
}

func insertChecked(t *testing.T, dbm *DBM, key, value string) {
	err := dbm.Insert([]byte(key), []byte(value))
	if err != nil {
		t.Errorf("Error on insert: %v", err)
	}
}

func TestDBM(t *testing.T) {
	// create a temp dir for test files
	tempdir, err := ioutil.TempDir("", "TestDBM")
	if err != nil {
		t.Fatalf("Couldn't create tempdir for test DB: %v", err)
	}
	defer os.RemoveAll(tempdir)

	// create a new DB in that temp dir
	dbm, err := Open(filepath.Join(tempdir, "test"))
	if err != nil {
		t.Fatalf("Couldn't open DB: %v", err)
	}
	defer dbm.Close()

	// check the empty database
	if dbm.Len() != 0 {
		t.Error("Empty DB should have no keys")
	}

	// insert some data
	{
		fill := [][]string{
			{"a", "alphabet"},
			{"b", "battlement"},
			{"c", "carnival"},
			{"d", "dinosaur"},
		}
		for _, keyValue := range fill {
			insertChecked(t, dbm, keyValue[0], keyValue[1])
		}
		if dbm.Len() != 4 {
			t.Errorf("DB should have 4 keys, but actually has %d", dbm.Len())
		}
	}

	// try to insert a key that already exists, which should fail
	{
		err := dbm.Insert([]byte("c"), []byte("contentment"))
		if err != nil && err != AlreadyExists {
			t.Errorf("Error on insert: %v", err)
		}
		if dbm.Len() != 4 {
			t.Errorf("DB should still have 4 keys, but actually has %d", dbm.Len())
		}
	}

	// replace a key that already exists
	{
		err := dbm.Replace([]byte("c"), []byte("contentment"))
		if err != nil {
			t.Errorf("Error on replace: %v", err)
		}
		if dbm.Len() != 4 {
			t.Errorf("DB should still have 4 keys, but actually has %d", dbm.Len())
		}
	}

	// delete a key
	{
		err := dbm.Delete([]byte("b"))
		if err != nil {
			t.Errorf("Error on delete: %v", err)
		}
		if dbm.Len() != 3 {
			t.Errorf("DB should have 3 keys, but actually has %d", dbm.Len())
		}
	}

	// delete a key that has never existed
	{
		err := dbm.Delete([]byte("x"))
		if err != nil && err != NotFound {
			t.Errorf("Error on delete: %#v", err)
		}
	}

	// delete a key that has already been deleted
	{
		err := dbm.Delete([]byte("b"))
		if err != nil && err != NotFound {
			t.Errorf("Error on delete: %#v", err)
		}
		if dbm.Len() != 3 {
			t.Errorf("DB should have 3 keys, but actually has %d", dbm.Len())
		}
	}
}
