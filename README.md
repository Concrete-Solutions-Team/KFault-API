# KFault API

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

