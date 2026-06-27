# KFault API

A Go-based backend service for real-time messaging and chat room management.

## Project Workflow & Development

This project uses Docker to manage infrastructure dependencies.

### Prerequisites
- Go 1.20+
- Docker & Docker Compose

### Development Workflow

1.  **Start Infrastructure**:
    Use the provided Makefile to spin up the required PostgreSQL database and Adminer (a web-based database management tool).
    ```bash
    make infra-up
    ```
    *This starts a PostgreSQL instance on port `5432` and Adminer on port `8080`.*

2.  **Run the Server**:
    Start the API server locally:
    ```bash
    make server-up
    ```

3.  **Tear Down**:
    Stop and remove the infrastructure containers:
    ```bash
    make infra-down
    ```

## API Documentation

### Authentication

| Method | Endpoint | Description | Auth Required |
| :--- | :--- | :--- | :--- |
| `POST` | `/auth/register` | Register a new user | No |
| `POST` | `/auth/login` | Authenticate user and set a session cookie | No |
| `GET` | `/auth/me` | Retrieve current user profile | Yes |
| `POST` | `/auth/logout` | Invalidate session and clear cookie | Yes |

### Rooms

| Method | Endpoint | Description | Auth Required |
| :--- | :--- | :--- | :--- |
| `GET` | `/rooms` | Get list of all available rooms | Yes |
| `POST` | `/rooms/create` | Create a new room | Yes |

### Messaging

| Method | Endpoint | Description | Auth Required |
| :--- | :--- | :--- | :--- |
| `GET` | `/chat` | Establish WebSocket connection for chat | Yes |
| `GET` | `/w` | Serves the WebSocket test UI | No |

### System

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `GET` | `/` | Health/Root check |
