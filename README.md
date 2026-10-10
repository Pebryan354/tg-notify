# TG Notify (Go Library)

Library Go ringan untuk monitoring dan pengiriman notifikasi ke Telegram (melalui Telegram Bot API). Dirancang khusus untuk alerting sistem ketika terjadi exception/error atau ketika terjadi kendala operasional (seperti integrasi dengan pihak ke-3 gagal).

Library ini **tidak membutuhkan database**, mudah diintegrasikan, dan mendukung fitur forum topic (**thread**) Telegram serta mode **asinkron** (background worker).

---

## Fitur Utama

- **Error & Exception Alert**: Otomatis memformat pesan error sistem disertai lokasi file, baris kode, dan fungsi pemanggil (`caller stack`).
- **Third-Party & Operational Alert**: Notifikasi khusus untuk kegagalan pihak ke-3 (misal: API Bank timeout, webhook gateway error, kuota habis).
- **Mendukung Telegram Thread / Topic**: Opsi `thread_id` untuk grup Telegram dengan mode topik/forum.
- **Dukungan Metadata**: Menambahkan data konteks tambahan (seperti `user_id`, `trx_code`, `request_id`, `endpoint`) dengan mudah.
- **Mode Asinkron**: Dapat berjalan di background goroutine sehingga tidak menghambat response time aplikasi utama jika koneksi Telegram lambat.
- **Auto Safe HTML & Truncation**: Karakter HTML otomatis di-escape agar tidak merusak formatting Telegram, serta pesan otomatis dipotong jika melebihi batas 4096 karakter Telegram.

---

## Instalasi

```bash
go get github.com/Pebryan354/tg-notify
```

---

## Cara Penggunaan

### 1. Inisialisasi

```go
package main

import (
	"log"

	tgnotify "github.com/Pebryan354/tg-notify"
)

func main() {
	botToken := "123456789:ABCdefGhIJKlmNoPQRsTUVwxyZ"
	chatID := "-1001234567890" // Chat ID group atau channel Telegram

	notifier, err := tgnotify.New(botToken, chatID,
		// Nama aplikasi & environment di header notifikasi
		tgnotify.WithAppName("Payment Service"),
		tgnotify.WithEnvironment("production"),

		// Opsional: ID thread jika grup menggunakan Forum Topics
		tgnotify.WithThreadID(42),

		// Opsional: Aktifkan mode asinkron (non-blocking)
		tgnotify.WithAsync(true),
		tgnotify.WithErrorHandler(func(err error) {
			log.Printf("Gagal kirim telegram: %v", err)
		}),
	)
	if err != nil {
		log.Fatalf("Inisialisasi tgnotify gagal: %v", err)
	}

	_ = notifier
}
```

---

### 2. Notifikasi Error / Exception Sistem

Gunakan `NotifyError` ketika menangani `err != nil` pada proses penting atau pada middleware recovery / panic:

```go
if err := processTransaction(); err != nil {
	// Notifikasi error akan otomatis menangkap file, baris, dan nama fungsi pemanggil
	_ = notifier.NotifyError(ctx, err,
		tgnotify.WithExtra(map[string]any{
			"user_id":  1042,
			"trx_code": "TRX-2026-999",
			"action":   "Create VA",
		}),
	)
}
```

**Tampilan di Telegram:**
```text
🚨 [Payment Service] [PRODUCTION] SYSTEM ERROR

Error:
connection refused to database master

Location:
• Function: services.ProcessTransaction
• File: /app/services/transaction.go:88

Metadata:
• action: Create VA
• trx_code: TRX-2026-999
• user_id: 1042

🕒 2026-10-10 19:15:00 WIB
```

---

### 3. Notifikasi Kegagalan Pihak ke-3 (Third-Party Alert)

Gunakan `NotifyAlert` untuk kondisi di mana sistem Anda berjalan normal, namun ada layanan eksternal yang gagal atau timeout:

```go
_ = notifier.NotifyAlert(ctx,
	"THIRD-PARTY API FAILURE",
	"Bank BCA Endpoint returned HTTP 504 Gateway Timeout during inquiry.",
	tgnotify.WithExtra(map[string]any{
		"partner":    "Bank BCA",
		"endpoint":   "https://api.bankbca.co.id/v1/inquiry",
		"attempt":    3,
		"time_spent": "5000ms",
	}),
	// Anda juga bisa mengarahkan alert ini ke thread khusus:
	tgnotify.PerThreadID(99),
)
```

**Tampilan di Telegram:**
```text
⚠️ [Payment Service] [PRODUCTION] THIRD-PARTY API FAILURE

Detail:
Bank BCA Endpoint returned HTTP 504 Gateway Timeout during inquiry.

Metadata:
• attempt: 3
• endpoint: https://api.bankbca.co.id/v1/inquiry
• partner: Bank BCA
• time_spent: 5000ms

🕒 2026-10-10 19:15:00 WIB
```

---

### 4. Kirim Pesan Kustom / Plain Text / HTML

Jika hanya ingin mengirim pesan biasa:

```go
// Kirim pesan teks biasa
_ = notifier.SendMessage(ctx, "Sistem daily backup berhasil dijalankan.")

// Kirim pesan dengan tag HTML custom
_ = notifier.SendHTML(ctx, "<b>DEPLOY INFO</b>: Version <code>v1.2.0</code> deployed successfully.")
```
