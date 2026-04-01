package kv

import (
	"encoding/binary"
	"fmt"
	"io"
)

const (
	opPut    byte = 1
	opDelete byte = 2
)

type logRecord struct {
	Op    byte
	Key   []byte
	Value []byte
}

// encode record ke byte
func encodeRecord(rec logRecord) []byte {
	keyLen := uint32(len(rec.Key))
	valLen := uint32(len(rec.Value))

	// 1 byte op + 4 byte keyLen + 4 byte valLen + key +  value
	buf := make([]byte, 1+4+4+len(rec.Key)+len(rec.Value))

	buf[0] = rec.Op
	binary.BigEndian.PutUint32(buf[1:5], keyLen)
	binary.BigEndian.PutUint32(buf[5:9], valLen)

	copy(buf[9:9+len(rec.Key)], rec.Key)
	copy(buf[9+len(rec.Key):], rec.Value)

	return buf
}

func decodeRecord(r io.Reader) (logRecord, error) {
	var header [9]byte

	_, err := io.ReadFull(r, header[:])
	if err != nil {
		return logRecord{}, err
	}
	op := header[0]
	keyLen := binary.BigEndian.Uint32(header[1:5])
	valLen := binary.BigEndian.Uint32(header[5:9])

	if op != opPut && op != opDelete {
		return logRecord{}, fmt.Errorf("invalid op: %d", op)
	}

	key := make([]byte, keyLen)
	_, err = io.ReadFull(r, key)
	if err != nil {
		return logRecord{}, err
	}

	value := make([]byte, valLen)
	_, err = io.ReadFull(r, value)
	if err != nil {
		return logRecord{}, err
	}
	return logRecord{
		Op:    op,
		Key:   key,
		Value: value,
	}, nil

}
