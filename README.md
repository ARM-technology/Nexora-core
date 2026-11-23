[README.md](https://github.com/user-attachments/files/35268856/README.md)
# Nexora Core

A system core for managing business-to-customer communication through WhatsApp. It combines a web interface with an API server for handling messages, employees, bookings, and automated replies.

> This was my first **full-stack** project, and I started it in **February 2025**. This repository contains an older version of the project and represents an important milestone in my learning and growth as a developer.

## Project Overview

Nexora Core aims to help businesses organize customer conversations and related workflows in one place. It includes a Go server, a prebuilt Flutter web interface, and integrations with WhatsApp and Google Gemini.

## Features

- Connect WhatsApp using a QR code and manage sessions.
- Send messages and maintain conversation history.
- Manage companies, employees, and some customer information.
- Configure automated replies and employee keywords.
- Manage appointment slots and bookings.
- AI configuration and experimental Gemini integration code.
- JWT-based login and real-time updates using Server-Sent Events (SSE).
- Manage subscription packages and usage limits.

These items describe functionality represented in this version of the source code; they do not guarantee that every route is complete or production-ready. Gemini integration code is present, but it is not connected to the bot's reply-generation flow in this version.

## Tech Stack

| Area | Technology |
| --- | --- |
| Server and API | Go, Gin |
| Database | PostgreSQL, pgx |
| Web interface | Flutter Web — prebuilt files |
| WhatsApp integration | whatsmeow |
| AI | Google Gemini API |
| Authentication and updates | JWT and SSE |

> **Frontend note:** `data/web` contains the compiled Flutter web build, not the Dart source files. The frontend cannot be modified or rebuilt from this repository alone.

## Project Structure

```text
Nexora-core/
├── main.go                         # Server entry point
├── APIs.go                         # API routes and frontend serving
├── APIWhatsApp.go                  # WhatsApp integration logic
├── AI_Employee.go                  # Gemini integration code
├── CyberJWT.go                     # JWT creation and validation
├── ServerSentEventsSSE.go          # SSE event streaming
├── useDataBaseSQL.go               # Database schema initialization
├── StructerAPIsJson.go             # Data types and database connection setup
├── Tools/Cryptography/Encryption/  # Encryption utilities
├── data/web/                       # Compiled Flutter web app
├── go.mod
└── System.env                      # Environment settings read by the app
```

## Run Locally

### Prerequisites

- Go **1.25.8** or a compatible version, as specified in `go.mod`.
- PostgreSQL and an accessible local database.
- A Gemini API key if you want to experiment with the AI-related code.
- A WhatsApp account to connect and try the related features.

### 1. Clone the repository and download dependencies

```bash
git clone https://github.com/ARM-technology/Nexora-core.git
cd Nexora-core
go mod download
```

### 2. Configure the database

Create a local PostgreSQL database, then update `DBUser`, `DBPassword`, `DBHost`, `DBPort`, and `DBName` in `StructerAPIsJson.go` for your environment. In this version, these values are defined in the source code rather than read from environment variables. **Do not commit real database credentials or passwords to the repository.**

The server attempts to initialize its tables at startup, so make sure the database user has permission to create them.

### 3. Configure environment variables

The application reads a file named `System.env` from the project root. Replace its values with your own local values, for example:

```dotenv
JWT_SECRET=replace_with_a_long_random_secret
GEMINI_API_KEY=replace_with_your_gemini_api_key
NAMEADMIAN=replace_with_your_admin_name
PASSADMIAN=replace_with_your_admin_password
MY_KMS_KEY_URI=replace_with_a_fake_kms_key_uri
```

Keep the variable names exactly as shown; `NAMEADMIAN` and `PASSADMIAN` are spelled this way in the source code. Do not publish this file with real credentials or keys.

`MY_KMS_KEY_URI` must use a `fake-kms://...` key URI accepted by Tink. To generate a key for **local testing only**, run the following Bash commands from the project directory, then put the printed URI in `System.env`:

```bash
tmpdir="$(mktemp -d)"
cat > "$tmpdir/fakekms-key.go" <<'EOF'
package main

import (
    "fmt"
    "log"

    "github.com/tink-crypto/tink-go/v2/testing/fakekms"
)

func main() {
    uri, err := fakekms.NewKeyURI()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(uri)
}
EOF

go run "$tmpdir/fakekms-key.go"
rm -r "$tmpdir"
```

> Tink's `fakekms` is intended for testing; it is not a secure production key-management system. Review and replace the encryption and key-management approach before any production deployment.

### 4. Start the server

```bash
go run .
```

If configuration succeeds, the web interface is served at:

```text
http://localhost:8080
```

The compiled frontend uses `localhost:8080` for API requests. When deploying elsewhere, you may need to change the API URL and rebuild the frontend from its Flutter source, which is not included in this repository.

## Example API Routes

| Method | Path | Purpose |
| --- | --- | --- |
| `POST` | `/login` | Log in |
| `POST` | `/register-company` | Register a company |
| `GET` | `/packages` | List packages |
| `GET` | `/events?company_id=1` | Receive company SSE events |
| `GET` | `/get-qr` | Get a WhatsApp connection QR code |
| `POST` | `/send-msg` | Send a message as an employee |
| `POST` | `/get-all-bookings` | List company bookings |
| `POST` | `/set-ai-config` | Configure AI settings |

Protected routes such as `/get-qr` and `/send-msg` require an authorization header in this format: `Authorization: Bearer <token>`. See `APIs.go` and the handler files for the available routes and request fields in this version.

## Security and Project Status

- This repository contains an older experimental version. It needs security review, testing, and updates before any production use.
- The repository version includes populated credentials in `System.env` and hard-coded database credentials in source code. If any of these values are real or still active, treat them as exposed: rotate or revoke them, remove them from version control, and use local environment configuration instead. Do not reuse the values from this snapshot.
- Review route authorization, data protection, password handling, and API key management before deploying. Do not rely on the included configuration as-is.
- This project uses the `whatsmeow` library for WhatsApp connectivity and is not an official WhatsApp or Meta product.

---

This project began as a learning experience in building a full-stack application, and it remains an important first step in my full-stack journey.
