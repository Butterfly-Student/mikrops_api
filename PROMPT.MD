Analisis proyek ini dengan menjadikan PLANING.md sebagai referensi awal dan dengan instruksi yang ada pada CLAUDE.MD, namun jangan menghasilkan ulang planning atau dokumen rekomendasi.

Tugasmu adalah bertindak sebagai software engineer yang langsung memperbaiki proyek, bukan sebagai reviewer.

Lakukan langkah-langkah berikut secara berurutan tahap demi tahap:

Validasi kesesuaian proyek terhadap PLANING.md
Periksa struktur direktori, pembagian domain, dan arsitektur.
Jika terdapat ketidaksesuaian, langsung sesuaikan implementasi proyek agar selaras dengan planning, tanpa meminta konfirmasi.

Perbaiki dan lengkapi testing per domain

Pastikan setiap domain memiliki file *_test.go yang menguji business logic utama.

Jika test belum ada, buatkan test-nya.

Jika test ada namun tidak relevan, refactor atau ganti dengan test yang benar.

Lengkapi HTTP / API testing

Audit direktori tests/http.

Jika belum ada atau belum lengkap, buat HTTP test untuk setiap endpoint penting (success, validation error, unauthorized, dan failure case).

Perbaiki fixture dan integration test

Pastikan tests/fixtures benar-benar digunakan oleh test.

Jika integration test belum ada atau tidak dapat dijalankan, buat atau perbaiki integration test (database, external service abstraction).

Sesuaikan integrasi third-party service

Proyek ini tidak menggunakan email.

Semua notifikasi harus menggunakan WhatsApp Gateway.

Semua pembayaran harus menggunakan Xendit.
Jika ditemukan:

logic email → hapus dan ganti dengan WhatsApp

coupling langsung ke service → buat abstraction + mock

kode tidak bisa di-test → refactor sampai testable

Gap fixing (bukan gap analysis)
Untuk setiap kekurangan, ketidaksesuaian, atau desain lemah yang kamu temukan:

identifikasi lokasi masalahnya (file, package, atau layer)

perbaiki atau refactor langsung pada kode

tambahkan file, test, atau abstraction yang kurang

hapus bagian yang tidak dipakai atau bertentangan dengan kebutuhan

Jangan berhenti di rekomendasi.
Jangan membuat ulang planning.
Hasil akhir harus berupa perubahan konkret pada proyek.