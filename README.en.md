# Eshak Pay 🫏

[🇷🇺 Русский](README.md) | 🇬🇧 English

# Eshak Pay 🫏

**Eshak Pay** is an educational electronic payment system built as a backend-focused project.

The project simulates the core operations of a simplified payment system: user registration, authentication, accounts, balance management, transfers between users, and transaction history.

> ⚠️ **Disclaimer:** Eshak Pay is an educational project and is not a real banking or payment service. It does not process real money.

## ✨ Features

- User registration and authentication
- Secure password hashing with Argon2id
- Server-side session management
- HTTP-only authentication cookies
- User accounts
- Balance management
- Transfers between users
- Idempotency protection for transfer requests
- Transaction history
- Account ID display and copying
- Light and dark UI themes
- REST API
- PostgreSQL database
- HTTPS deployment

## 🏗 Architecture

The backend follows a layered architecture:

```text
Browser / Client
       │
       ▼
    HTTP API
       │
       ▼
   Handlers
       │
       ▼
    Services
       │
       ▼
 Repositories
       │
       ▼
   PostgreSQL
```

The project uses a single Go application with PostgreSQL rather than a microservice architecture.

## 🛠 Tech Stack

### Backend

- Go
- `net/http`
- PostgreSQL
- `pgx`
- Raw SQL
- Clean / layered architecture

### Frontend

- HTML
- CSS
- JavaScript
- Manrope font
- Light / dark themes

### Security

- Argon2id password hashing
- Opaque server-side sessions
- SHA-256 session token hashing
- HTTP-only cookies
- HTTPS
- Idempotency keys
- PostgreSQL constraints and transactions

## 💳 Transfers

A transfer between two users is processed as a transactional operation.

The system uses an idempotency key to prevent accidental repeated processing of the same request.

Financial transactions are stored separately from transfer requests, allowing the system to distinguish between successfully executed operations and rejected requests.

## 📊 Transaction History

Users can view their previous financial operations through the transaction history.

The history includes incoming and outgoing transfers as well as the initial account credit.

## 🔐 Authentication

Eshak Pay uses server-side sessions instead of JWT.

After successful authentication, the server creates an opaque session token and stores its hash in PostgreSQL. The token is sent to the browser through an HTTP-only cookie.

Passwords are never stored in plain text and are hashed using Argon2id.

## 🗄 Database

The project uses PostgreSQL with raw SQL queries.

Main entities include:

```text
users
sessions
accounts
cards
transactions
postings
transfer_requests
```

Financial operations use signed postings:

- debit — negative amount
- credit — positive amount

## 📁 Project Structure

```text
eshak-pay/
├── cmd/
│   └── server/
├── internal/
│   ├── config/
│   ├── database/
│   ├── domain/
│   ├── repository/
│   ├── security/
│   └── service/
├── web/
│   └── static/
│       ├── css/
│       ├── js/
│       └── index.html
├── migrations/
├── go.mod
└── README.md
```

## 🚀 Running Locally

### Requirements

- Go 1.22+
- PostgreSQL

### 1. Clone the repository

```bash
git clone https://github.com/rabotyaga-ctrl-v/eshak-pay.git
cd eshak-pay
```

### 2. Configure environment variables

Create a `.env` file:

```env
DB_HOST=localhost
DB_PORT=5432
DB_NAME=eshak_pay
DB_USER=eshak_app
DB_PASSWORD=your_password
DB_SSLMODE=disable
```

### 3. Create the database

Create a PostgreSQL database and apply the project's SQL migrations/schema.

### 4. Install dependencies

```bash
go mod tidy
```

### 5. Run the server

```bash
go run ./cmd/server
```

The application will be available locally at:

```text
http://localhost:8080
```

## 🌐 Production

The project is deployed on a Linux VPS using:

```text
Internet
   │
   ▼
 HTTPS
   │
   ▼
 Nginx
   │
   ▼
 Go application
   │
   ▼
 PostgreSQL
```

The production application is available at:

**https://eshakpay.ru**

## 📱 Mobile Application

A native Android client is planned as the next stage of the project.

The mobile application will use the existing backend API and provide:

- authentication
- personal account
- balance
- transfers
- transaction history

The Android client will be developed as a separate client application rather than replacing the existing backend.

## 🗺 Roadmap

### v1.0 — Completed

- [x] Registration
- [x] Authentication
- [x] Sessions
- [x] Accounts
- [x] Balance
- [x] Transfers
- [x] Transaction history
- [x] Light / dark themes
- [x] Production deployment

### v1.1+

- [ ] Android application
- [ ] Extended security testing
- [ ] Load testing
- [ ] Performance analysis
- [ ] Further system improvements

## 🎓 Educational Purpose

Eshak Pay was created as a practical project for studying backend development, databases, security, transactional operations, and system architecture.

The project may also be used as an experimental platform for further research into:

- transaction security
- concurrent operations
- system performance
- load testing
- reliability of payment operations

## 📄 License

This project is intended primarily for educational and research purposes.