# Ledger Project - Financial Engine in Go

This is a high-performance Ledger system developed in Go, focusing on **data consistency**, **transaction atomicity**, and **concurrency safety**. The project applies advanced architectural patterns to ensure that financial movements are immutable and auditable.

## ⚙️ Architecture & Design Principles
The project follows a **Layered Architecture** approach (paving the way for Hexagonal Architecture), ensuring strict separation of concerns:

* **Domain Driven:** Business rules are encapsulated within rich models (`Account`, `Transaction`), not scattered across the code.
* **Repository Pattern:** Abstracts the persistence layer, allowing the business logic to remain agnostic of the underlying SQL implementation.
* **Service Layer:** Orchestrates complex operations. It manages database transactions (ACID), ensuring that a fund transfer is either fully completed or completely rolled back.
* **Strong Typing:** Usage of `uuid.UUID` and `int64` (for currency) to prevent precision errors and type mismatches.

## 🛠️ Tech Stack
* **Language:** Go (Golang) 1.2x+
* **Database:** PostgreSQL
* **Driver:** `pgx/v5` (Chosen for its superior performance and native type support compared to `lib/pq`)
* **Security:** Environment variables (`.env`) for sensitive data and Strict Transport Security logic.

## 🚀 How to Run

### Prerequisites
* Go installed
* PostgreSQL running (Docker or Local)
* `.env` file configured in the root directory

### Installation & Execution
1.  **Clone the repository:**
    ```bash
    git clone [https://github.com/SEU_USUARIO/ledger-project.git](https://github.com/SEU_USUARIO/ledger-project.git)
    cd ledger-project
    ```

2.  **Install dependencies:**
    ```bash
    go mod tidy
    ```

3.  **Run the Server & Seed Data:**
    This command starts the HTTP server and populates the database with test accounts.
    ```bash
    go run cmd/ledger/main.go cmd/ledger/seed.go
    ```
    *Output should indicate: `Ledger API online at http://localhost:8080`*

## 🧪 Testing the API
You can verify the system behavior using `curl`. Replace the UUIDs below with the ones generated in your terminal during the seed process.

**Sample Request (Transfer):**
```bash
curl -X POST http://localhost:8080/transfer \
-H "Content-Type: application/json" \
-d '{
    "account_from_id": "YOUR_SENDER_UUID_HERE",
    "account_to_id": "YOUR_RECEIVER_UUID_HERE",
    "amount": 5000,
    "description": "Integration Test Transfer"
}'