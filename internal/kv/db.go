package kv

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
)

type DB struct {
	mu   sync.RWMutex
	path string
	file *os.File
	data map[string][]byte
}

// open membuka database, membuat file kalau belum ada, lalu me reply semua log memory
func Open(path string) (*DB, error) {
	fp, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, fmt.Errorf("open db file:%w", err)
	}

	db := &DB{
		path: path,
		file: fp,
		data: make(map[string][]byte),
	}

	if err := db.load(); err != nil {
		fp.Close()
		return nil, err
	}

	// penting: setelah load, posisikan  pointer ke akhir file
	if _, err := fp.Seek(0, io.SeekEnd); err != nil {
		fp.Close()
		return nil, fmt.Errorf("seek end: %w", err)
	}
	return db, nil
}

// load-> me-reply seluruh log ke state memory
func (db *DB) load() error {
	if _, err := db.file.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("seek start: %w", err)
	}
	for {
		rec, err := decodeRecord(db.file)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			if errors.Is(err, io.ErrUnexpectedEOF) {
				// nanti di fase lebih lanjut kita bahas recovery lebih proper.
				// untuk sekarang anggap log korup bila record terakhir terpotong
				return fmt.Errorf("corrupted log: %w ", err)
			}
			return fmt.Errorf("decode record: %w ", err)
		}
		key := string(rec.Key)
		switch rec.Op {
		case opPut:
			valCopy := append([]byte(nil), rec.Value...)
			db.data[key] = valCopy
		case opDelete:
			delete(db.data, key)
		default:
			return fmt.Errorf("unknown op: %d", rec.Op)
		}
	}
	return nil
}

// put:  menyimpan/mengupdate key
func (db *DB) Put(key string, value []byte) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	rec := logRecord{
		Op:    opPut,
		Key:   []byte(key),
		Value: value,
	}

	if err := db.appendRecord(rec); err != nil {
		return err
	}

	db.data[key] = append([]byte(nil), value...)
	return nil
}

func (db *DB) Get(key string) ([]byte, bool) {
	db.mu.Lock()
	defer db.mu.Unlock()

	val, ok := db.data[key]
	if !ok {
		return nil, false
	}

	return append([]byte(nil), val...), true
}

func (db *DB) Delete(key string) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	rec := logRecord{
		Op:    opDelete,
		Key:   []byte(key),
		Value: nil,
	}

	if err := db.appendRecord(rec); err != nil {
		return err
	}

	delete(db.data, key)
	return nil
}

func (db *DB) Close() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if db.file == nil {
		return nil
	}

	err := db.file.Close()
	db.file = nil
	return err
}

// appendRecord menulis record ke file dan fsync agar durable.
func (db *DB) appendRecord(rec logRecord) error {
	buf := encodeRecord(rec)

	_, err := db.file.Write(buf)
	if err != nil {
		return fmt.Errorf("write log: %w", err)
	}

	if err := db.file.Sync(); err != nil {
		return fmt.Errorf("fsync log: %w", err)
	}
	return nil

}

// DebugDump untuk melihat isi state memory saat belajar/debug.
func (db *DB) DebugDump() string {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var out bytes.Buffer
	for k, v := range db.data {
		out.WriteString(fmt.Sprintf("%s=%s\n", k, string(v)))
	}
	return out.String()
}
