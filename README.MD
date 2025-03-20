# Muzz gRPC Service

## Overview
This gRPC service implements the core functionality of Muzz's Explore Service, allowing users to interact by liking or passing on other profiles. The service is built using **Go**, **gRPC**, and **MySQL**.

## Features
- **ListLikedYou**: Lists all users who have liked the recipient.
- **ListNewLikedYou**: Lists users who liked the recipient but were not liked back.
- **CountLikedYou**: Counts the number of users who have liked the recipient.
- **PutDecision**: Records a like or pass decision and updates mutual like status.

---

## Prerequisites
Ensure you have the following installed:
- **Go (>=1.20)**
- **Docker & Docker Compose**
- **MySQL**
- **gRPCurl (for testing gRPC endpoints)**

---

## Setup & Running the Service
### 1️⃣ Clone the Repository
```sh
git clone https://github.com/Dev-AustinPeter/explore-service-muzz.git
cd explore-service-muzz
```

### 2️⃣ Start MySQL & gRPC Using Docker
```sh
docker-compose up --build
```
This runs MySQL with the necessary database and tables pre-configured.
And gRPC server

<!--
### 3️⃣ Build & Run the gRPC Server
```sh
go build -o explore_service
./explore_service
```
The server will start on **port 50051**.

--- -->

## gRPC API Endpoints & Testing
Use `grpcurl` to test the gRPC endpoints.

### **1️⃣ ListLikedYou**
```sh
grpcurl -plaintext -d '{"recipient_user_id": "user2"}' localhost:50051 ExploreService/ListLikedYou
```

### **2️⃣ ListNewLikedYou**
```sh
grpcurl -plaintext -d '{"recipient_user_id": "user2"}' localhost:50051 ExploreService/ListNewLikedYou
```

### **3️⃣ CountLikedYou**
```sh
grpcurl -plaintext -d '{"recipient_user_id": "user2"}' localhost:50051 ExploreService/CountLikedYou
```

### **4️⃣ PutDecision (Like/Pass)**
```sh
grpcurl -plaintext -d '{"actor_user_id": "user1", "recipient_user_id": "user2", "liked_recipient": true}' localhost:50051 ExploreService/PutDecision
```

---

## Running Tests
### **Unit Tests**
```sh
go test -tags unit -v
```
or
```sh
make test
```
This runs all unit tests using `sqlmock` to mock database interactions.

---

## Database Schema
The `decisions` table is structured as follows:
```sql
CREATE DATABASE IF NOT EXISTS muzz;

USE muzz;

CREATE TABLE IF NOT EXISTS decisions (
    id INT AUTO_INCREMENT PRIMARY KEY,
    actor_user_id VARCHAR(36) NOT NULL,
    recipient_user_id VARCHAR(36) NOT NULL,
    liked BOOLEAN NOT NULL,
    unix_timestamp BIGINT DEFAULT (UNIX_TIMESTAMP()),
    UNIQUE KEY unique_decision (actor_user_id, recipient_user_id),
    INDEX (recipient_user_id),
    INDEX (actor_user_id)
);
```

---

## Troubleshooting
### **Database Connection Issues**
Ensure MySQL is running:
```sh
docker ps | grep mysql
```
If not, restart MySQL:
```sh
docker-compose restart
```

### **gRPC Connection Issues**
Ensure the service is running and listening on **port 50051**:
```sh
netstat -an | grep 50051
```
If not, restart the service:
```sh
./explore_service
```

---


