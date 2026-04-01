package pager

import (
	"os"
	"testing"
)

const (
	dbName = "pager_test.db"
)

func TestPagerWrite(t *testing.T) {
	_ = os.Remove(dbName)

	p, err := Open(dbName)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		_ = p.Close()
		_ = os.Remove(dbName)

	}()

	pageID, err := p.AllocatePage()
	if err != nil {
		t.Fatal(err)
	}

	if pageID != 0 {
		t.Fatalf("expected first page = 0 got = %d", pageID)
	}

	data := make([]byte, PageSize)
	copy(data, []byte("hello world"))

	err = p.WritePage(pageID, data)
	if err != nil {
		t.Fatal(err)
	}

	read, err := p.ReadPage(pageID)
	if err != nil {
		t.Fatalf("data mismatch: got:%q", string(read[:11]))
	}
}

func TestPagerAllocateTwoPages(t *testing.T) {
	_ = os.Remove(dbName)

	p, err := Open(dbName)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		_ = p.Close()
		_ = os.Remove(dbName)

	}()

	page0, err := p.AllocatePage()
	if err != nil {
		t.Fatal(err)
	}

	page1, err := p.AllocatePage()
	if err != nil {
		t.Fatal(err)
	}

	if page0 != 0 {
		t.Fatalf("expected first page = 0 got = %d", page0)
	}

	if page1 != 1 {
		t.Fatalf("expected first page = 1 got = %d", page1)
	}

	n, err := p.NumPages()
	if err != nil {
		t.Fatal(err)
	}

	if n != 2 {
		t.Fatalf("expected num pages =2. got %d", n)

	}

	info, err := os.Stat(dbName)
	if err != nil {
		t.Fatal(err)
	}

	if info.Size() != 2*PageSize {
		t.Fatalf("expected file size = %d, got = %d", 2*PageSize, info.Size())
	}

}
