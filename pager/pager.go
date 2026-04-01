package pager

import (
	"fmt"
	"os"
)

// Artinya:
// page 0 = byte 0 sampai 4095
// page 1 = byte 4096 sampai 8191
// page 2 = byte 8192 sampai 12287
// dst
const PageSize = 4096

type Pager struct {
	file *os.File
}

func Open(path string) (*Pager, error) {
	fp, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644) // file boleh dibaca, file boleh ditulis ,kalau belum ada, buat file baru
	if err != nil {
		return nil, fmt.Errorf("open pager file: %w", err)
	}

	return &Pager{
		file: fp,
	}, nil
}

func (p *Pager) Close() error {
	if p.file == nil {
		return nil
	}
	return p.file.Close()
}

func (p *Pager) ReadPage(pageID uint64) ([]byte, error) {
	// 	Kalau mau baca page ke-3:
	// offset = 3 × 4096
	// berarti mulai baca dari byte ke-12288
	offset := int64(pageID) * PageSize

	buf := make([]byte, PageSize)

	// 	ReadAt bagus karena:
	// tidak tergantung posisi pointer file saat ini
	// langsung baca dari offset tertentu
	// cocok untuk random access
	n, err := p.file.ReadAt(buf, offset)
	if err != nil {
		return nil, fmt.Errorf("read page %d: %w", pageID, err)
	}

	// 	Kalau yang terbaca cuma 100 byte:
	// byte 0..99 = data dari file
	// byte 100..4095 = kita isi nol
	// Kenapa ini berguna?
	// Karena nanti struktur page mengharapkan panjang page selalu konsisten.
	if n < PageSize {
		for i := n; i < PageSize; i++ {
			buf[i] = 0
		}
	}

	return buf, nil
}

func (p *Pager) WritePage(pageID uint64, data []byte) error {
	// 	Ini wajib:
	// Kenapa?
	// Karena satu page harus selalu ukuran tetap.
	// Kalau kita izinkan:
	// 300 byte
	// 5000 byte
	// maka konsep page rusak total.
	if len(data) != PageSize {
		return fmt.Errorf("invalid page size: got=%d want=%d", len(data), PageSize)
	}

	// 	pageID 0 → offset 0
	// pageID 1 → offset 4096
	// pageID 2 → offset 8192
	offset := int64(pageID) * PageSize

	_, err := p.file.WriteAt(data, offset)
	if err != nil {
		return fmt.Errorf("write page %d: %w", pageID, err)
	}

	// 	Ini memastikan write didorong ke storage dengan lebih aman.
	// Tanpa Sync():
	// data bisa masih di page cache OS
	// crash/listrik mati bisa bikin data belum benar-benar persisted
	if err := p.file.Sync(); err != nil {
		return fmt.Errorf("fsync page %d: %w", pageID, err)
	}

	return nil
}

func (p *Pager) AllocatePage() (uint64, error) {
	info, err := p.file.Stat()
	if err != nil {
		return 0, fmt.Errorf("start pager file: %w", err)
	}

	size := info.Size()

	if size%PageSize != 0 {
		return 0, fmt.Errorf("file size is not aligned to page size: %d", size)
	}

	pageID := uint64(size / PageSize)

	// Kenapa tidak cukup return pageID saja?
	// Karena kita ingin:
	// file benar-benar bertambah
	// page itu benar-benar ada secara fisik
	empty := make([]byte, PageSize)

	if err := p.WritePage(pageID, empty); err != nil {
		return 0, err
	}

	return pageID, nil
}

func (p *Pager) NumPages() (uint64, error) {
	info, err := p.file.Stat()
	if err != nil {
		return 0, fmt.Errorf("stat pager file: %w", err)
	}

	size := info.Size()

	if size%PageSize != 0 {
		return 0, fmt.Errorf("file size is not aligned to page size: %d", size)
	}

	return uint64(size / PageSize), nil
}
