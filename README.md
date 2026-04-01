# my-own-db

Project belajar membuat database sederhana dengan Go.

Saat ini repo ini berisi dua komponen utama:

- `internal/kv`: key-value store sederhana berbasis append-only log.
- `pager`: abstraction untuk membaca dan menulis page berukuran tetap `4096` byte.

## Fitur Saat Ini

- `put`, `get`, `del` untuk operasi key-value dasar.
- Persistence ke file lewat append-only log.
- Replay log saat startup untuk membangun state in-memory.
- `fsync` setelah write agar perubahan lebih aman saat crash.
- Pager sederhana dengan alokasi page baru, read page, write page, dan hitung jumlah page.

## Struktur Project

```text
.
├── cmd/
│   └── main.go
├── internal/
│   ├── btree/
│   │   └── node.go
│   └── kv/
│       ├── codec.go
│       ├── db.go
│       └── db_test.go
├── pager/
│   ├── pager.go
│   └── pager_test.go
├── data/
│   └── my.db
└── go.mod
```

## Cara Kerja `internal/kv`

Database menyimpan data dalam dua bentuk:

- Di memori: `map[string][]byte` untuk akses cepat.
- Di disk: file log append-only untuk durability.

Format record log:

```text
+---------+----------+----------+---------+-----------+
| 1B op   | 4B kLen  | 4B vLen  | key     | value     |
+---------+----------+----------+---------+-----------+
```

Nilai `op`:

- `1` = `put`
- `2` = `delete`

Saat database dibuka:

1. File database dibuka atau dibuat jika belum ada.
2. Semua record dibaca ulang dari awal file.
3. State akhir direkonstruksi ke memory map.

Saat `Put` atau `Delete` dipanggil:

1. Record baru di-append ke file.
2. File di-`Sync()`.
3. State in-memory diperbarui.

## Cara Kerja `pager`

Package `pager` mengelola file sebagai kumpulan page berukuran tetap:

- `PageSize = 4096`
- `page 0` dimulai dari byte `0`
- `page 1` dimulai dari byte `4096`
- dst

Method yang tersedia:

- `Open(path string)`
- `ReadPage(pageID uint64)`
- `WritePage(pageID uint64, data []byte)`
- `AllocatePage()`
- `NumPages()`

Package ini cocok sebagai fondasi untuk layer storage yang lebih rendah, misalnya B-Tree atau slotted page.

## Menjalankan CLI

CLI saat ini ada di `cmd/main.go` dan menggunakan path database relatif:

```go
dbPath := "../data/my.db"
```

Karena itu, jalankan command dari direktori `cmd`:

```bash
cd cmd
go run . put name andi
go run . get name
go run . del name
go run . dump
```

Contoh output:

```bash
$ go run . put name andi
ok

$ go run . get name
andi

$ go run . del name
ok
```

Jika key tidak ada:

```bash
$ go run . get unknown
nil
```

## Menjalankan Test

Dari root project:

```bash
go test ./...
```

Status verifikasi saat README ini dibuat:

- `internal/kv`: lulus
- `pager`: lulus

## Keterbatasan Saat Ini

- Semua data aktif tetap disimpan di memory map, jadi belum cocok untuk data besar.
- Belum ada compaction atau segment merge, jadi file log akan terus membesar.
- `dump` memakai iterasi map Go, jadi urutan output tidak stabil.
- `Get` memakai `Lock()` bukan `RLock()`, sehingga concurrency read belum optimal.
- Penanganan tail corruption masih sederhana: record terakhir yang terpotong dianggap error.
- `internal/btree` masih tahap awal dan belum dipakai oleh layer `kv`.
- CLI masih bergantung pada working directory karena path database hardcoded relatif.

## Ide Pengembangan Berikutnya

- Tambahkan compaction untuk membersihkan record lama.
- Ganti storage in-memory penuh dengan struktur berbasis page.
- Lanjutkan implementasi B-Tree di atas pager.
- Tambahkan format metadata file dan recovery yang lebih robust.
- Perbaiki CLI agar path database bisa di-pass lewat flag.

## Lisensi

Belum ada lisensi yang didefinisikan di repository ini.
