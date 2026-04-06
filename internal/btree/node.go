package btree

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const (
	BNODE_NODE = 1 //internal node menyimpan key + child pointer
	BNODE_LEAF = 1 //leaf menyimpan key + value
)

// | type | nkeys | pointers | offsets | key-values |
// |  2B  |  2B   | nkeys*8B | nkeys*2B|    ...     |

const (
	HEADER             = 4 // Karena header berisi: type 2 byte nkeys 2 byte
	BTREE_PAGE_SIZE    = 4096
	BTREE_MAX_SIZE     = 1000
	BTREE_MAX_VAL_SIZE = 3000
)

type BNode struct {
	data []byte
}

// Helper untuk membuat node kosong.
func NewBNode(size int) BNode {
	return BNode{data: make([]byte, size)}
}

// cek isi raw bytes
// kirim ke pager
// debug
func (node BNode) Bytes() []byte {
	return node.data
}

// Membaca 2 byte pertama sebagai tipe node.
func (node BNode) btype() uint16 {
	return binary.LittleEndian.Uint16(node.data[0:2])
}

// Membaca jumlah key dalam node.
// ada berapa pointer
// ada berapa offset
// valid index sampai mana
func (node BNode) nkeys() uint16 {
	return binary.LittleEndian.Uint16(node.data[2:4])
}

// Menulis header ke byte 0..3.
// Kapan dipakai?
// Saat kita membuat node baru, misalnya: leaf dengan 2 key
// internal node dengan 3 key
func (node BNode) setHeader(btype uint16, nkeys uint16) {
	binary.LittleEndian.PutUint16(node.data[0:2], btype)
	binary.LittleEndian.PutUint16(node.data[2:4], nkeys)
}

// Mengambil pointer child ke-idx. Kenapa 8 byte? Karena pointer page kita uint64.
// Rumus posisi:
// Header 4 byte, lalu pointer list mulai setelah itu.
// Kalau:
// idx=0 → posisi 4
// idx=1 → posisi 12
// idx=2 → posisi 20
func (node BNode) getPtr(idx uint16) uint64 {
	assert(idx < node.nkeys(), "setPtr: idx out of range")
	pos := HEADER + 8*idx
	return binary.LittleEndian.Uint64(node.data[pos : pos+8])
}

func (node BNode) setPtr(idx uint16, val uint64) {
	assert(idx < node.nkeys(), "setPtr: idx out of range")
	pos := HEADER + 8*idx
	binary.LittleEndian.PutUint64(node.data[pos:pos+8], val)
}

//	KV pair panjangnya variabel.
//
// Contoh:
// key1 = "a" , value1 = "123"
// key2 = "kota", value2 = "pekanbaru"
// Kalau mau langsung lompat ke KV ke-2, kita butuh tahu posisi awalnya.
// Itulah fungsi offset list. Buku menjelaskan bahwa offset relatif terhadap posisi KV pertama,
// offset pertama tidak disimpan karena selalu 0,
// dan offset terakhir dipakai untuk mengetahui ukuran node.
func offsetPos(node BNode, idx uint16) uint16 {
	assert(1 <= idx && idx <= node.nkeys(), "offsetPost: idx out of range")
	return HEADER + 8*node.nkeys() + 2*(idx-1)
}

// Mengambil offset relatif untuk KV ke-idx.
// Kenapa idx==0 return 0?
// Karena KV pertama selalu mulai dari offset 0 relatif terhadap area KV.
func (node BNode) getOffset(idx uint16) uint16 {
	if idx == 0 {
		return 0
	}
	return binary.LittleEndian.Uint16(node.data[offsetPos(node, idx) : offsetPos(node, idx)+2])
}

// Menulis offset list.
// Nanti kapan dipakai?
// Saat kita append atau copy KV ke node baru.
func (node BNode) setOffset(idx uint16, offset uint16) {
	assert(1 <= idx && idx <= node.nkeys(), "offsetPost: idx out of range")
	pos := offsetPos(node, idx)
	binary.LittleEndian.PutUint16(node.data[pos:pos+2], offset)
}

// Menghitung posisi byte tempat KV ke-idx dimulai.
// Rumusnya:
// lewat header
// lewat semua pointer
// lewat semua offset
// lalu tambah offset relatif KV ke-idx
// Kenapa offset area pakai 2 * nkeys()?
// Karena buku menyimpan offset list untuk membantu locate KV dengan cepat,
// dan node size bisa dihitung dari posisi akhir KV terakhir.
func (node BNode) kvPos(idx uint16) uint16 {
	assert(1 <= node.nkeys(), "kvPos: idx out of range")
	return HEADER + 8*node.nkeys() + 2*node.nkeys() + node.getOffset(idx)
}

// Mengambil key ke-idx.
// Kenapa pos+4?
// Karena di awal KV pair ada:
// klen 2 byte
// vlen 2 byte
// baru setelah itu data key.
func (node BNode) getKey(idx uint16) []byte {
	assert(1 <= node.nkeys(), "getKey: idx out of range")
	pos := node.kvPos(idx)
	klen := binary.LittleEndian.Uint16(node.data[pos : pos+2])
	return node.data[pos+4 : pos+4+klen]
}

// Mengambil value ke-idx.
// Kenapa mulai dari pos+4+klen?
// Karena value berada setelah:
// header kecil KV (4 byte)
// key bytes
func (node BNode) getVal(idx uint16) []byte {
	assert(1 <= node.nkeys(), "getVal: idx out of range")
	pos := node.kvPos(idx)
	klen := binary.LittleEndian.Uint16(node.data[pos : pos+2])
	vlen := binary.LittleEndian.Uint16(node.data[pos+2 : pos+4])
	return node.data[pos+4+klen : pos+4+klen+vlen]
}

// Menghitung ukuran node yang benar-benar terpakai.
// Kenapa penting?
// Karena nanti:
// kalau nbytes() <= 4096 → node masih muat 1 page
// kalau nbytes() > 4096 → harus split
func (node BNode) nbytes() uint16 {
	return node.kvPos(node.nkeys())
}

func assert(ok bool, msg string) {
	if !ok {
		panic(fmt.Sprintf("assert failed: %s", msg))
	}
}

// fungsi yang akan menentukan:
// key baru harus dimasukkan di posisi mana di dalam node
// Misalnya node berisi key:
// [a][c][f]
// kita mau insert "d"-> kita harus tahu: masuk di antara c dan f Jadi hasilnya:
// [a][c][d][f]
// nodeLookupLE = Lookup Less or Equal
// index:   0   1   2
// keys :  [a] [c] [f]
// Cari "d"
// "a" ≤ "d" → ya
// "c" ≤ "d" → ya
// "f" ≤ "d" → tidak
// hasil: index = 1
func nodeLookupLE(node BNode, key []byte) uint16 {
	// berapa banyak key dalam node
	n := node.nkeys()

	var i uint16

	// kita scan dari kiri ke kanan
	for i = 0; i < n; i++ {
		// ini ambil key dari byte-level node (yang tadi kita buat)
		k := node.getKey(i)

		// Hasil:
		// < 0 → k < key
		// 0 → k == key
		// > 0 → k > key
		cmp := bytes.Compare(k, key)

		if cmp > 0 {
			break
		}
	}

	if i == 0 {
		return 0
	}

	return i - 1
}

func nodeAppendKV(new BNode, idx uint16, ptr uint64, key []byte, value []byte) {
	// 1. set pointer
	new.setPtr(idx, ptr)

	// cari posisi awal KV ke idx
	pos := int(new.kvPos(idx))

	// tulis key length dan value length
	binary.LittleEndian.PutUint16(new.data[pos:pos+2], uint16(len(key)))
	binary.LittleEndian.PutUint16(new.data[pos+2:pos+4], uint16(len(value)))

	// copy key dan value
	copy(new.data[pos+4:], key)
	copy(new.data[pos+4+len(key):], value)

	// hitung ukuran KV
	kvSize := uint16(4 + len(key) + len(value))

	new.setOffset(idx+1, new.getOffset(idx)+kvSize)

}
