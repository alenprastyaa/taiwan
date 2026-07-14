# University Agency Dashboard

Dashboard operasional untuk Taiwan Education Consulting dengan role owner, staff, dan client.

## Tech Stack

- Go + chi
- a-h/templ component renderer
- Vanilla JS kecil untuk navigasi SPA-like
- CSS statis yang di-embed ke binary Go
- PWA manifest + service worker
- Konfigurasi 12 factor melalui `.env`
- Database adapter siap Postgres/MySQL/MariaDB
- Tidak membutuhkan Node.js untuk build atau runtime production

## Development

```bash
go build -buildvcs=false -o ./bin/university-agency.exe ./cmd/server
./bin/university-agency.exe
```

Untuk hot reload dengan Air:

```bash
go run -buildvcs=false github.com/air-verse/air@v1.61.7 -c .air.toml
```

`DATABASE_DRIVER` wajib diisi (`postgres`, `mysql`, atau `mariadb`). Aplikasi akan gagal start jika DB tidak bisa dihubungi; database harus diprovision lebih dulu. Saat terhubung, aplikasi memakai SQL repository dan auto-migrate schema. Demo seed hanya aktif saat `APP_ENV=development`; production DB kosong harus dibootstrap eksplisit.

## Akun Demo Lokal

- Owner: `owner` / `owner12345`
- Staff: `staff` / `staff12345`
- Client: `student` / `student12345`

Registrasi mandiri hanya membuat akun client/student. Owner dan staff disediakan dari seed repository sampai adapter database produksi diaktifkan.
