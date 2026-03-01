# pkg/gowa — WhatsApp Gateway Client

Package `gowa` adalah reusable Go client untuk berinteraksi dengan [Gowa WhatsApp API](https://github.com/aldinokemal/go-whatsapp-web-multidevice) (WhatsApp MultiDevice API).

Package ini mendukung:
- Multi-device management
- Login via QR code atau pairing code
- Kirim pesan teks, gambar, file, dan video
- Manajemen grup (list, cari, info)
- Manajemen device (tambah, hapus, status)
- Informasi user WhatsApp

---

## Konfigurasi

Konfigurasi diambil dari environment variables. Tambahkan ke file `.env`:

```env
# URL server Gowa API (default: http://localhost:3000)
GOWA_BASE_URL=http://localhost:3000

# Basic Auth credentials (wajib)
GOWA_USERNAME=admin
GOWA_PASSWORD=admin

# Device ID default (opsional)
GOWA_DEVICE_ID=my-device-id

# Timeout request HTTP dalam detik (default: 30)
GOWA_TIMEOUT=30
```

---

## Instalasi & Inisialisasi

### Menggunakan environment variables (direkomendasikan)

```go
import "go-template/pkg/gowa"

client, err := gowa.NewFromEnv()
if err != nil {
    log.Fatal(err)
}
```

### Menggunakan konfigurasi manual

```go
import (
    "time"
    "go-template/pkg/gowa"
)

cfg := &gowa.Config{
    BaseURL:  "http://localhost:3000",
    Username: "admin",
    Password: "admin",
    DeviceID: "my-device-id",
    Timeout:  30 * time.Second,
}

client := gowa.New(cfg)
```

### Menggunakan device ID berbeda per request

```go
// Buat client dengan device ID berbeda
clientWithDevice := client.WithDeviceID("other-device-id")
```

---

## App / Login

### Login via QR Code

```go
ctx := context.Background()

// Login menggunakan device default
resp, err := client.Login(ctx)
if err != nil {
    log.Fatal(err)
}

fmt.Println("QR Code URL:", resp.Results.QRLink)
fmt.Println("QR Duration:", resp.Results.QRDuration, "seconds")
```

### Login via Pairing Code

```go
// Login dengan nomor telepon (format: 628xxxxxxxxx)
resp, err := client.LoginWithCode(ctx, "628912344551")
if err != nil {
    log.Fatal(err)
}

fmt.Println("Pairing Code:", resp.Results.PairCode)
```

### Cek Status Koneksi

```go
status, err := client.GetStatus(ctx)
if err != nil {
    log.Fatal(err)
}

if status.Results.IsLoggedIn {
    fmt.Println("Device sudah login dan terhubung")
} else {
    fmt.Println("Device belum login")
}
```

### Logout

```go
resp, err := client.Logout(ctx)
if err != nil {
    log.Fatal(err)
}
fmt.Println(resp.Message)
```

### Reconnect

```go
resp, err := client.Reconnect(ctx)
if err != nil {
    log.Fatal(err)
}
fmt.Println(resp.Message)
```

---

## Kirim Pesan

### Kirim Pesan Teks ke Individual

```go
// Format nomor: 6289685028129@s.whatsapp.net
phone := gowa.FormatPhoneNumber("6289685028129")

resp, err := client.SendTextMessage(ctx, phone, "Halo dari MikroTik ISP!")
if err != nil {
    log.Fatal(err)
}

fmt.Println("Message ID:", resp.Results.MessageID)
```

### Kirim Pesan ke Grup

```go
// Format group JID: 120363347168689807@g.us
groupJID := gowa.FormatGroupJID("120363347168689807")

resp, err := client.SendGroupMessage(ctx, groupJID, "Notifikasi untuk semua member!")
if err != nil {
    log.Fatal(err)
}
```

### Kirim Pesan dengan Opsi Lengkap

```go
resp, err := client.SendMessage(ctx, gowa.SendMessageRequest{
    Phone:   gowa.FormatPhoneNumber("6289685028129"),
    Message: "Halo! Ini pesan dengan mention.",
    Mentions: []string{"628123456789", "@everyone"},
    Duration: 3600, // Disappearing message 1 jam
})
```

### Kirim Gambar dari URL

```go
resp, err := client.SendImageFromURL(
    ctx,
    gowa.FormatPhoneNumber("6289685028129"),
    "https://example.com/invoice.png",
    "Invoice bulan Februari 2026",
    "", // deviceID kosong = gunakan default
)
```

### Kirim Gambar dari File Lokal

```go
resp, err := client.SendImageFromFile(
    ctx,
    gowa.FormatPhoneNumber("6289685028129"),
    "/path/to/invoice.png",
    "Invoice bulan Februari 2026",
    "",
)
```

### Kirim File dari URL

```go
resp, err := client.SendFileFromURL(
    ctx,
    gowa.FormatPhoneNumber("6289685028129"),
    "https://example.com/invoice.pdf",
    "Invoice PDF",
    "",
)
```

---

## Manajemen Grup

### Dapatkan Semua Grup

```go
groups, err := client.GetMyGroups(ctx)
if err != nil {
    log.Fatal(err)
}

for _, group := range groups {
    fmt.Printf("Grup: %s | JID: %s | Members: %d\n",
        group.Name, group.JID, len(group.Participants))
}
```

### Cari Grup Berdasarkan Nama

```go
// Pencarian case-insensitive, partial match
group, err := client.FindGroupByName(ctx, "ISP Notifications")
if err != nil {
    log.Fatal(err)
}

fmt.Println("Group JID:", group.JID)
fmt.Println("Group Name:", group.Name)
```

### Dapatkan JID Grup Berdasarkan Nama

```go
// Shortcut untuk mendapatkan JID langsung
groupJID, err := client.GetGroupJIDByName(ctx, "ISP Notifications")
if err != nil {
    log.Fatal(err)
}

// Langsung kirim pesan ke grup
resp, err := client.SendGroupMessage(ctx, groupJID, "Pesan notifikasi!")
```

### Dapatkan Info Grup

```go
info, err := client.GetGroupInfo(ctx, "120363347168689807@g.us")
if err != nil {
    log.Fatal(err)
}
```

### Dapatkan Link Undangan Grup

```go
linkResp, err := client.GetGroupInviteLink(ctx, "120363347168689807@g.us")
if err != nil {
    log.Fatal(err)
}

fmt.Println("Invite Link:", linkResp.Results.InviteLink)
```

---

## Manajemen Device

### List Semua Device

```go
devices, err := client.ListDevices(ctx)
if err != nil {
    log.Fatal(err)
}

for _, device := range devices {
    fmt.Printf("Device: %s | State: %s | Phone: %s\n",
        device.ID, device.State, device.PhoneNumber)
}
```

### Tambah Device Baru

```go
// Dengan custom device ID
device, err := client.AddDevice(ctx, "my-custom-device")
if err != nil {
    log.Fatal(err)
}

// Tanpa device ID (server akan generate)
device, err = client.AddDevice(ctx, "")
```

### Cek Status Device

```go
status, err := client.GetDeviceStatus(ctx, "my-device-id")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Connected: %v, Logged In: %v\n",
    status.IsConnected, status.IsLoggedIn)
```

### Cek Apakah Device Terhubung

```go
isConnected, err := client.IsDeviceConnected(ctx, "my-device-id")
if err != nil {
    log.Fatal(err)
}

if isConnected {
    fmt.Println("Device siap digunakan")
}
```

### Login Device dengan QR Code

```go
resp, err := client.LoginDevice(ctx, "my-device-id")
if err != nil {
    log.Fatal(err)
}

fmt.Println("Scan QR:", resp.Results.QRLink)
```

### Login Device dengan Pairing Code

```go
resp, err := client.LoginDeviceWithCode(ctx, "my-device-id", "628912344551")
if err != nil {
    log.Fatal(err)
}

fmt.Println("Pairing Code:", resp.Results.PairCode)
```

### Logout Device

```go
resp, err := client.LogoutDevice(ctx, "my-device-id")
if err != nil {
    log.Fatal(err)
}
```

### Hapus Device

```go
resp, err := client.RemoveDevice(ctx, "my-device-id")
if err != nil {
    log.Fatal(err)
}
```

---

## Informasi User

### Cek Apakah Nomor Ada di WhatsApp

```go
isOnWhatsApp, err := client.CheckUser(ctx, "628912344551")
if err != nil {
    log.Fatal(err)
}

if isOnWhatsApp {
    fmt.Println("Nomor terdaftar di WhatsApp")
}
```

### Dapatkan Info User

```go
info, err := client.GetUserInfo(ctx, gowa.FormatPhoneNumber("6289685028129"))
if err != nil {
    log.Fatal(err)
}
```

### Dapatkan Daftar Kontak

```go
contacts, err := client.GetMyContacts(ctx)
if err != nil {
    log.Fatal(err)
}

for _, contact := range contacts {
    fmt.Println(contact)
}
```

---

## Helper Functions

```go
// Format nomor telepon ke WhatsApp JID
phone := gowa.FormatPhoneNumber("6289685028129")
// Output: 6289685028129@s.whatsapp.net

// Format group ID ke WhatsApp group JID
groupJID := gowa.FormatGroupJID("120363347168689807")
// Output: 120363347168689807@g.us
```

---

## Contoh Penggunaan untuk Notifikasi ISP

### Kirim Notifikasi Invoice ke Customer

```go
func SendInvoiceNotification(ctx context.Context, customerPhone, invoiceNumber string, amount float64) error {
    client, err := gowa.NewFromEnv()
    if err != nil {
        return err
    }

    message := fmt.Sprintf(
        "📋 *Invoice Baru*\n\nNomor: %s\nJumlah: Rp %.0f\n\nSilakan lakukan pembayaran sebelum tanggal jatuh tempo.",
        invoiceNumber, amount,
    )

    phone := gowa.FormatPhoneNumber(customerPhone)
    _, err = client.SendTextMessage(ctx, phone, message)
    return err
}
```

### Kirim Notifikasi ke Grup Admin

```go
func NotifyAdminGroup(ctx context.Context, message string) error {
    client, err := gowa.NewFromEnv()
    if err != nil {
        return err
    }

    groupJID, err := client.GetGroupJIDByName(ctx, "Admin ISP")
    if err != nil {
        return err
    }

    _, err = client.SendGroupMessage(ctx, groupJID, message)
    return err
}
```

### Kirim Notifikasi Isolasi Customer

```go
func SendIsolationNotification(ctx context.Context, customerPhone, customerName string) error {
    client, err := gowa.NewFromEnv()
    if err != nil {
        return err
    }

    message := fmt.Sprintf(
        "⚠️ *Pemberitahuan Isolasi*\n\nYth. %s,\n\nLayanan internet Anda telah diisolasi karena tagihan yang belum dibayar.\n\nSilakan hubungi kami untuk informasi lebih lanjut.",
        customerName,
    )

    phone := gowa.FormatPhoneNumber(customerPhone)
    _, err = client.SendTextMessage(ctx, phone, message)
    return err
}
```

---

## Error Handling

Semua method mengembalikan error yang deskriptif:

```go
resp, err := client.SendTextMessage(ctx, phone, message)
if err != nil {
    // Error bisa berupa:
    // - "send message request failed: ..." (network error)
    // - "send message failed: request failed with status 400: ..." (API error)
    // - "send message failed: request failed with status 500: ..." (server error)
    log.Printf("Failed to send message: %v", err)
    return err
}
```

---

## Struktur Package

```
pkg/gowa/
├── README.md       # Dokumentasi ini
├── config.go       # Konfigurasi dari environment variables
├── models.go       # Struct request/response
├── client.go       # HTTP client dengan Basic Auth
├── app.go          # Login, logout, reconnect, status
├── send.go         # Kirim pesan teks, gambar, file, video
├── group.go        # Manajemen grup WhatsApp
├── device.go       # Manajemen multi-device
└── user.go         # Informasi user WhatsApp
```
