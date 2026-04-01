package kv

import (
	"os"
	"testing"
)

func TestBasicPutGet(t *testing.T) {
	os.Remove("test.db")

	db, err := Open("test.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	err = db.Put("name", []byte("andi"))
	if err != nil {
		t.Fatal(err)
	}

	val, ok := db.Get("name")
	if !ok {
		t.Fatal("key not found")
	}

	if string(val) != "andi" {
		t.Fatalf("expected andi, got %s", val)
	}

}

func TestPersistence(t *testing.T) {
	os.Remove("test.db")

	db, _ := Open("test.db")
	db.Put("city", []byte("pekanbaru"))
	db.Close()

	db2, _ := Open("test.db")
	defer db2.Close()

	val, ok := db2.Get("city")
	if !ok {
		t.Fatal("data not persisted")
	}

	if string(val) != "pekanbaru" {
		t.Fatalf("expected pekanbaru, got %s", val)
	}
}

func TestCrashRecovery(t *testing.T) {
	os.Remove("test.db")

	db, _ := Open("test.db")

	db.Put("a", []byte("1"))
	db.Put("b", []byte("2"))

	// simulasi crash: kita tidak Close()

	db2, err := Open("test.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db2.Close()

	val, ok := db2.Get("a")
	if !ok || string(val) != "1" {
		t.Fatal("crash recovery failed for key a")
	}

	val, ok = db2.Get("b")
	if !ok || string(val) != "2" {
		t.Fatal("crash recovery failed for key b")
	}
}

func TestCorruptedTail(t *testing.T) {
	os.Remove("test.db")

	db, _ := Open("test.db")
	db.Put("x", []byte("123"))
	db.Close()

	// buka file dan potong sebagian (simulate corruption)
	f, _ := os.OpenFile("test.db", os.O_RDWR, 0644)
	f.Truncate(10) // potong sembarangan
	f.Close()

	// reopen database
	db2, err := Open("test.db")
	if err == nil {
		defer db2.Close()
	}
}
