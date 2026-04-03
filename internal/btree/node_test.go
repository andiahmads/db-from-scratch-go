package btree

import (
	"encoding/binary"
	"testing"
)

func TestNodeHeader(t *testing.T) {
	node := NewBNode(BTREE_PAGE_SIZE)
	node.setHeader(BNODE_LEAF, 3)

	if node.btype() != BNODE_LEAF {
		t.Fatalf("expected type=%d got=%d", BNODE_LEAF, node.btype())
	}

	if node.nkeys() != 3 {
		t.Fatalf("expected nkeys=3 got=%d", node.nkeys())
	}
}

func TestNodePointer(t *testing.T) {
	node := NewBNode(BTREE_PAGE_SIZE)
	node.setHeader(BNODE_LEAF, 3)

	node.setPtr(0, 100)
	node.setPtr(1, 200)
	node.setPtr(2, 300)

	if node.getPtr(0) != 100 {
		t.Fatalf("expected ptr0=100 got=%d", node.getPtr(0))
	}

	if node.getPtr(1) != 200 {
		t.Fatalf("expected ptr1=200 got=%d", node.getPtr(1))
	}

	if node.getPtr(2) != 300 {
		t.Fatalf("expected ptr2=300 got=%d", node.getPtr(2))
	}

}

func TestNodeqOffsetAndKv(t *testing.T) {
	node := NewBNode(BTREE_PAGE_SIZE)
	node.setHeader(BNODE_LEAF, 2)

	// kita isi pointer area walau leaf, biar layout tetap konsisten
	node.setPtr(0, 0)
	node.setPtr(1, 0)

	// KV 0: key = "a", val="123"
	pos0 := node.kvPos(0)
	binary.LittleEndian.PutUint16(node.data[pos0:pos0+2], uint16(len("a")))
	binary.LittleEndian.PutUint16(node.data[pos0+2:pos0+4], uint16(len("123")))

	copy(node.data[pos0+4:], []byte("a"))
	copy(node.data[pos0+5:], []byte("123"))
	node.setOffset(1, 4+1+3) // header kv + key + val = 8

	// KV 1: key = "bb", val="xyz"
	pos1 := node.kvPos(1)
	binary.LittleEndian.PutUint16(node.data[pos1:pos1+2], uint16(len("bb")))
	binary.LittleEndian.PutUint16(node.data[pos1+2:pos1+4], uint16(len("xyz")))

	copy(node.data[pos1+4:], []byte("bb"))
	copy(node.data[pos1+6:], []byte("xyz"))
	node.setOffset(2, node.getOffset(1)+4+2+3)

	if string(node.getKey(0)) != "a" {
		t.Fatalf("expected key0=a got=%q", string(node.getKey(0)))
	}

	if string(node.getVal(0)) != "123" {
		t.Fatalf("expected val0=123 got=%q", string(node.getVal(0)))
	}

	if string(node.getKey(1)) != "bb" {
		t.Fatalf("expected key1=a got=%s", string(node.getKey(1)))
	}

	if string(node.getVal(1)) != "xyz" {
		t.Fatalf("expected val1=a got=%q", string(node.getVal(1)))
	}

}

func TestNodeBytesUsed(t *testing.T) {
	node := NewBNode(BTREE_PAGE_SIZE)
	node.setHeader(BNODE_LEAF, 1)
	node.setPtr(0, 0)

	pos := node.kvPos(0)
	binary.LittleEndian.PutUint16(node.data[pos:pos+2], 1)
	binary.LittleEndian.PutUint16(node.data[pos+2:pos+4], 3)

	copy(node.data[pos+4:], []byte("a"))
	copy(node.data[pos+5:], []byte("123"))

	node.setOffset(1, 4+1+3)

	want := uint16(pos + 4 + 1 + 3)
	if node.nbytes() != want {
		t.Fatalf("expected nbytes=%d, got=%d", want, node.nbytes())
	}

}
