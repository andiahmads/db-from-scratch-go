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
	// 	key = "a" → panjang = 1
	// value = "123" → panjang = 3
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

func TestNodeLookUpLE(t *testing.T) {
	node := NewBNode(BTREE_PAGE_SIZE)
	node.setHeader(BNODE_LEAF, 3)

	// isi pointer dummy
	node.setPtr(0, 0)
	node.setPtr(1, 0)
	node.setPtr(2, 0)

	// KV :0 = "a"
	pos := node.kvPos(0)
	writeKV(node, pos, "a", "1")
	node.setOffset(1, kvSize("a", "1"))

	// KV :1 = "c"
	pos = node.kvPos(1)
	writeKV(node, pos, "c", "1")
	node.setOffset(2, node.getOffset(1)+kvSize("c", "1"))

	// KV :2 = "f"
	pos = node.kvPos(2)
	writeKV(node, pos, "f", "1")

	tests := []struct {
		key  string
		want uint16
	}{
		{"a", 0},
		{"b", 0},
		{"c", 1},
		{"d", 1},
		{"f", 2},
		{"z", 2},
	}

	for _, tt := range tests {
		got := nodeLookupLE(node, []byte(tt.key))
		if got != tt.want {
			t.Fatalf("key=%s want=%d got=%d", tt.key, tt.want, got)
		}
	}
}

// helper biar test lebiwriteKVh rapi
// index: 100 101 102 103 104 105 106 107
// data : [klen][vlen][ a ][ 1 ][ 2 ][ 3 ]
func writeKV(node BNode, pos uint16, k, v string) {
	key := []byte(k)
	val := []byte(v)

	p := int(pos)

	binary.LittleEndian.PutUint16(node.data[p:p+2], uint16(len(key)))
	binary.LittleEndian.PutUint16(node.data[p+2:p+4], uint16(len(val)))

	copy(node.data[p+4:], key)
	copy(node.data[p+4+len(key):], val)
}

func kvSize(k, v string) uint16 {
	return uint16(4 + len(k) + len(v))
}

func TestNodeAppendKV(t *testing.T) {
	node := NewBNode(BTREE_PAGE_SIZE)
	node.setHeader(BNODE_LEAF, 2)

	// pointer dummy
	node.setPtr(0, 0)
	node.setPtr(1, 0)

	nodeAppendKV(node, 0, 0, []byte("a"), []byte("123"))
	nodeAppendKV(node, 1, 0, []byte("bb"), []byte("xyz"))

	if string(node.getKey(0)) != "a" {
		t.Fatalf("expected key0=a got=%q", string(node.getKey(0)))
	}
	if string(node.getVal(0)) != "123" {
		t.Fatalf("expected val0=123 got=%q", string(node.getVal(0)))
	}
	if string(node.getKey(1)) != "bb" {
		t.Fatalf("expected key1=bb got=%q", string(node.getKey(1)))
	}
	if string(node.getVal(1)) != "xyz" {
		t.Fatalf("expected val1=xyz got=%q", string(node.getVal(1)))
	}

	// klen = 2 byte, vlen = 2 byte, key = 1 byte, val = 3 byte
	if node.getOffset(1) != 8 {
		t.Fatalf("expected offset1=8 got=%d", node.getOffset(1))
	}

	wantOffset2 := uint16(8 + 4 + 2 + 3) // KV0 + KV1
	if node.getOffset(2) != wantOffset2 {
		t.Fatalf("expected offset2=%d got=%d", wantOffset2, node.getOffset(2))
	}
}

func TestNodeAppendRange(t *testing.T) {
	old := NewBNode(BTREE_PAGE_SIZE)
	old.setHeader(BNODE_LEAF, 3)

	old.setPtr(0, 0)
	old.setPtr(1, 0)
	old.setPtr(2, 0)

	nodeAppendKV(old, 0, 0, []byte("a"), []byte("1"))
	nodeAppendKV(old, 1, 0, []byte("c"), []byte("2"))
	nodeAppendKV(old, 2, 0, []byte("f"), []byte("3"))

	new := NewBNode(BTREE_PAGE_SIZE)
	new.setHeader(BNODE_LEAF, 2)

	new.setPtr(0, 0)
	new.setPtr(1, 0)

	nodeAppendRange(new, old, 0, 1, 2)
	if string(new.getKey(0)) != "c" {
		t.Fatalf("expected key0=c got=%q", string(new.getKey(0)))
	}

	if string(new.getVal(0)) != "2" {
		t.Fatalf("expected val0=2 got=%q", string(new.getVal(0)))
	}

	if string(new.getKey(1)) != "f" {
		t.Fatalf("expected key1=f got=%q", string(new.getKey(1)))
	}

	if string(new.getVal(1)) != "3" {
		t.Fatalf("expected val1=3 got=%q", string(new.getVal(1)))
	}
}

func TestLeafInsert(t *testing.T) {
	old := NewBNode(BTREE_PAGE_SIZE)
	old.setHeader(BNODE_LEAF, 3)

	old.setPtr(0, 0)
	old.setPtr(1, 0)
	old.setPtr(2, 0)

	nodeAppendKV(old, 0, 0, []byte("a"), []byte("1"))
	nodeAppendKV(old, 1, 0, []byte("c"), []byte("2"))
	nodeAppendKV(old, 2, 0, []byte("f"), []byte("3"))

	new := leafInsert(old, 2, []byte("d"), []byte("X"))
	if new.nkeys() != 4 {
		t.Fatalf("expected nkeys=4 got=%d", new.nkeys())
	}

	wantKeys := []string{"a", "c", "d", "f"}
	wantVals := []string{"1", "2", "X", "3"}

	for i := range 4 {
		if string(new.getKey(uint16(i))) != wantKeys[i] {
			t.Fatalf("key[%d] want=%q got=%q", i, wantKeys[i], string(new.getKey(uint16(i))))
		}

		if string(new.getVal(uint16(i))) != wantVals[i] {
			t.Fatalf("val[%d] want=%q got=%q", i, wantVals[i], string(new.getVal(uint16(i))))
		}

	}
}
