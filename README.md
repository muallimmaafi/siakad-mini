# SIAKAD Mini API

RESTful API back end untuk SIAKAD Mini, layanan akademik sederhana yang mengelola
data mahasiswa, mata kuliah, dan Kartu Rencana Studi (KRS). Dibuat untuk UTS
Pemrograman Backend Lanjut.

## Tech Stack

- Go
- Fiber v2 (web framework)
- GORM + PostgreSQL
- JWT (`golang-jwt/jwt/v5`) untuk autentikasi token
- bcrypt untuk hash password
- `go-playground/validator` untuk validasi request

## Fitur

- 2 role: `admin` dan `mahasiswa`
- Autentikasi Bearer token di semua endpoint kecuali login
- Rate limiting login: maksimal 5 kali gagal per menit per IP (429)
- Soft delete mahasiswa (tidak muncul di daftar dan tidak bisa login)
- Pengambilan mata kuliah dalam satu transaction dengan row locking
- Format response JSON seragam, error 500 tanpa stack trace

## Struktur Folder

```
cmd/main.go          entry point API
cmd/seed/main.go     seeder data awal
config/              koneksi database
database/            migrasi (AutoMigrate)
internal/model/      model: User, Student, Course, Enrollment
internal/repository/ akses database
internal/service/    logika bisnis
internal/handler/    HTTP handler
internal/middleware/ autentikasi dan otorisasi role
pkg/                 helper response, JWT, validasi
```

## Cara Menjalankan

Prasyarat: Go, PostgreSQL, Git.

```bash
# 1. Clone
git clone https://github.com/muallimmaafi/siakad-mini.git
cd siakad-mini

# 2. Buat database
psql -U postgres -c "CREATE DATABASE siakad_mini;"

# 3. Siapkan konfigurasi (lalu isi password DB dan JWT_SECRET)
cp .env.example .env        # PowerShell: Copy-Item .env.example .env

# 4. Unduh dependency
go mod download

# 5. Jalankan seeder (membuat tabel + data awal)
go run ./cmd/seed

# 6. Jalankan API
go run cmd/main.go
```

API berjalan di `http://localhost:3000`. Migrasi tabel dijalankan otomatis saat
API atau seeder dimulai.

## Data Seeder

| Data | Jumlah | Keterangan |
|---|---|---|
| Admin | 1 | `admin@siakad.test` / `admin12345` |
| Mahasiswa | 20 | email `mhs01@siakad.test` s.d. `mhs20@siakad.test`, NIM `187221000001` s.d. `187221000020`, **password awal = NIM** |
| Mata kuliah | 10 | `TI101` s.d. `TI502` (TI502 berkuota 2, untuk uji kuota penuh) |

Seeder aman dijalankan berulang kali (data yang sudah ada dilewati).

## Daftar Endpoint

Base URL: `/api/v1`

| No | Method | Endpoint | Akses | Status sukses |
|---|---|---|---|---|
| 1 | POST | `/auth/login` | Publik | 200 |
| 2 | GET | `/auth/me` | Semua role | 200 |
| 3 | GET | `/students` | Admin | 200 |
| 4 | POST | `/students` | Admin | 201 |
| 5 | GET | `/students/{id}` | Admin, mahasiswa (data sendiri) | 200 |
| 6 | PUT | `/students/{id}` | Admin | 200 |
| 7 | DELETE | `/students/{id}` | Admin | 204 |
| 8 | GET | `/courses` | Semua role | 200 |
| 9 | POST | `/enrollments` | Mahasiswa | 201 |
| 10 | DELETE | `/enrollments/{id}` | Mahasiswa (milik sendiri) | 204 |

Query parameter:

- `GET /students`: `page`, `per_page` (maks. 50), `prodi`, `angkatan`, `search`
  (nim atau nama), `sort` (`nama` atau `-ipk_terakhir`)
- `GET /students/{id}`: `tahun_akademik` (opsional, menyaring total SKS)
- `GET /courses`: `semester`, `search` (kode atau nama), `available=true`,
  `tahun_akademik` (opsional, kuota dihitung per tahun akademik)

## Business Rule

1. Batas SKS per tahun akademik berdasarkan IPK terakhir: IPK >= 3,00 maksimal 24
   SKS; 2,50 sampai 2,99 maksimal 21 SKS; di bawah 2,50 maksimal 18 SKS
   (melebihi batas: 422, pesan menyebut sisa SKS).
2. Mata kuliah yang sama tidak dapat diambil dua kali pada tahun akademik yang
   sama (409).
3. Mata kuliah yang kuotanya penuh tidak dapat diambil (422).
4. Mahasiswa hanya dapat mengakses dan mengubah KRS miliknya sendiri (403).

## Format Response

Sukses:

```json
{
  "success": true,
  "message": "Data mahasiswa berhasil diambil",
  "data": [],
  "meta": { "current_page": 1, "per_page": 10, "total": 20, "last_page": 2 }
}
```

Error validasi (422):

```json
{
  "success": false,
  "message": "Validasi gagal",
  "errors": { "nim": ["NIM sudah terdaftar"] }
}
```

Status code yang dipakai: 200, 201, 204, 401, 403, 404, 409, 422, 429, 500.

## Catatan Desain

- Kuota dan batas SKS dihitung per `tahun_akademik`, karena mata kuliah yang sama
  dibuka kembali setiap semester.
- `POST /enrollments` memakai satu transaction: baris mahasiswa dan baris mata
  kuliah dikunci dengan `SELECT ... FOR UPDATE` sebelum cek duplikasi, kuota,
  dan batas SKS, sehingga permintaan bersamaan tidak bisa melampaui kuota atau
  batas SKS.
- Rate limiter login menyimpan hitungan di memori server dan membatasi per IP.
- `POST /students` membuat `users` dan `students` dalam satu transaction;
  password awal adalah NIM yang di-hash.

## Contoh Penggunaan

```bash
# Login
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@siakad.test","password":"admin12345"}'

# Pakai token
curl http://localhost:3000/api/v1/students \
  -H "Authorization: Bearer <access_token>"
```

## Pembuat

Muallim Maafi Ahmad, NIM 434241122, TI-C8, D4 Teknik Informatika, Universitas Airlangga.