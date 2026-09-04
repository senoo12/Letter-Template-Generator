# Letter Template Generator

Aplikasi web untuk generate surat massal secara otomatis dari template Word (.docx) dan data Excel (.xlsx), hasil akhirnya berupa banyak PDF yang otomatis dibungkus jadi satu file ZIP.

## Fitur

- Pengerjaan tanpa database, semua file diproses sementara lalu dihapus otomatis, tidak ada data pengguna yang tersimpan permanen (lihat bagian Privasi & Keamanan Data)
- Upload template Word (.docx) berisi placeholder `{{nama_field}}`
- Upload data Excel (.xlsx) dengan header kolom yang sama dengan nama placeholder
- Satu baris di Excel = satu surat yang di-generate
- Auto-convert setiap surat ke PDF
- Dukungan **gambar**: kalau value di Excel berupa link gambar (contoh: link tanda tangan), otomatis di-download dan ditanam sebagai gambar asli di suratnya, bukan sekadar teks link
- Output berupa satu file `.zip` berisi semua PDF yang di-generate
- Penamaan file otomatis berdasarkan kolom yang kamu tentukan sebagai "Identifier Kolom" (contoh: `Nama_Siswa`)

## Privasi & Keamanan Data

Proyek ini sengaja dirancang tanpa database. Semua file yang di-upload (template Word & data Excel) hanya diproses sementara di memori/temporary folder server selama request berlangsung, lalu langsung dihapus otomatis setelah surat selesai di-generate dan file ZIP hasilnya siap diunduh.

### Implikasinya:

- Tidak ada data pengguna (isi surat, data pribadi di Excel, dsb) yang tersimpan permanen di server
- Tidak ada riwayat/log dokumen yang bisa diakses ulang atau dilihat pihak lain
- File hasil generate di folder temporary server bersifat sementara, begitu proses request selesai, file input dihapus; file ZIP hasil unduhan juga sebaiknya dianggap sementara (server tidak menjamin penyimpanan jangka panjang)
- Cocok dipakai untuk data sensitif (data karyawan, calon karyawan, dsb) karena tidak ada jejak penyimpanan permanen di sisi server

Pendekatan ini dipilih supaya alur kerja terasa seperti "dijalankan di browser" dari sisi pengguna: upload, proses, unduh, selesai, sehingga tidak perlu khawatir data tersimpan di infrastruktur pihak ketiga.
> **Catatan:** karena tidak ada database, aplikasi ini juga tidak mendukung fitur seperti riwayat generate, multi-user session, atau resume proses yang terputus. Kalau ke depannya butuh fitur itu, perlu ditambahkan lapisan penyimpanan (database/object storage) secara terpisah, dengan trade-off privasi di atas perlu dipertimbangkan ulang.

## Cara Kerja Singkat

1. Di template Word, taruh placeholder dengan format `{{nama_kolom}}`, contoh:
   ```
   Kepada Yth. {{nama_penerima}}
   di {{kota}}
   ```
2. Di Excel, buat kolom dengan **nama header persis sama** dengan nama placeholder (`nama_penerima`, `kota`, dst).
3. Satu baris Excel akan menghasilkan satu surat.
4. Kalau value suatu kolom berupa link gambar (URL langsung ke file gambar), sistem otomatis mendeteksi dan menanamnya sebagai gambar di posisi placeholder tersebut, cocok untuk kolom tanda tangan (`ttd`), logo, foto, dll.
5. Semua surat hasil generate (dalam bentuk PDF) dibungkus jadi satu `.zip` untuk diunduh.

---

## Instalasi & Menjalankan Backend

### Prasyarat

- [Go](https://go.dev/dl/) versi **1.24** atau lebih baru
- **LibreOffice** terpasang di sistem (dipakai untuk convert docx → PDF headless)
  - Ubuntu/Debian: `sudo apt install libreoffice`
  - macOS: `brew install --cask libreoffice`
  - Windows: unduh installer dari [libreoffice.org](https://www.libreoffice.org/download/download/)
- Pastikan perintah `soffice` bisa dipanggil dari terminal (biasanya otomatis masuk ke PATH setelah instalasi)

### Langkah instalasi

```bash
# 1. Masuk ke folder backend
cd backend

# 2. Download dependency Go
go mod download
go mod tidy

# 3. Jalankan server
go run cmd/server/main.go
```

Server akan berjalan di `http://localhost:8080` secara default (bisa diubah lewat environment variable `PORT`).

Cek server sudah hidup:
```bash
curl http://localhost:8080/health
# -> {"status":"ok"}
```

### Build untuk production

```bash
cd backend
go build -o letter-generator ./cmd/server
./letter-generator
```

### Environment Variables

| Variable | Default | Keterangan |
|---|---|---|
| `PORT` | `8080` | Port yang dipakai server |

### Endpoint API

| Method | Endpoint | Deskripsi |
|---|---|---|
| `GET` | `/health` | Cek status server |
| `POST` | `/generate` | Upload template + excel, generate surat (multipart form: `template`, `excel`, `base_file_field`) |
| `GET` | `/downloads/:filename` | Download hasil generate (.zip) |

---

## Instalasi & Menjalankan Frontend

### Prasyarat

- [Node.js](https://nodejs.org/) versi 18 ke atas
- Package manager: `npm`, `pnpm`, atau `yarn` (sesuaikan dengan lockfile yang ada di project)

### Langkah instalasi

```bash
# 1. Masuk ke folder frontend
cd frontend

# 2. Install dependency
npm install
# atau: pnpm install / yarn install

# 3. Jalankan development server
npm run dev
```

Aplikasi akan berjalan di `http://localhost:5173` secara default.

### Konfigurasi koneksi ke backend

Pastikan frontend menunjuk ke URL backend yang benar. Cek/buat file environment di root folder `frontend`:

```bash
# copy .env.example
cp .env.example .env # for mac, linux, and git
copy .env.example .env # for windows

# frontend/.env
PUBLIC_API_BASE_URL=http://localhost:8080
```

Sesuaikan nama variable ini dengan yang dipakai di `$lib/services/letter` (service yang menangani pemanggilan API generate & download).

### File contoh template

Taruh file contoh Template Word dan Data Excel di folder `frontend/static/templates/` supaya tombol "Unduh contoh template" di halaman utama berfungsi:

```
frontend/static/templates/
├── Template_Surat.docx
└── Template_Data_Surat.xlsx
```

### Build untuk production

```bash
cd frontend
npm run build
npm run preview   # untuk preview hasil build secara lokal
```

---

## Menjalankan Full Stack (Development)

Jalankan backend dan frontend di dua terminal terpisah:

```bash
# Terminal 1 - Backend
cd backend
go run cmd/server/main.go

# Terminal 2 - Frontend
cd frontend
npm run dev
```

Lalu buka `http://localhost:5173` di browser.

---

## Troubleshooting

**Error "soffice: command not found" saat generate**
LibreOffice belum terpasang atau belum masuk PATH. Install LibreOffice dan pastikan bisa dipanggil dari terminal dengan `soffice --version`.

**Gambar dari Excel tidak muncul di surat**
- Pastikan link yang dipakai adalah link **langsung** ke file gambar (bukan link ke halaman artikel/website)
- Cek log server: sistem mencatat detail alasan gagal-nya di baris berawalan `[template]`
- Beberapa CDN memblokir hotlink; sistem sudah menyertakan header `Referer` untuk menyiasati ini, tapi tidak semua CDN bisa ditembus

**CORS error di frontend**
Backend sudah mengizinkan semua origin (`Access-Control-Allow-Origin: *`) secara default untuk kebutuhan development. Untuk production, sebaiknya batasi ke domain frontend yang spesifik di `cmd/server/main.go`.

## Lisensi

Proyek ini dilisensikan di bawah MIT License: bebas dipakai, dimodifikasi, dan didistribusikan ulang oleh siapa saja, termasuk untuk keperluan komersial, selama notice copyright & lisensi aslinya tetap disertakan.