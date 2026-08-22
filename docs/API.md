# Recess API Specification — Version 1 (`/api/v1`)

> Production REST API specification for Recess, covering authentication, identity, player profiles, game metadata, and lobby rooms.

---

## 1. Global Conventions

### 1.1 Base URL
- Local: `http://localhost:8080/api/v1`
- Production: `https://api.recess.game/api/v1`

### 1.2 Headers
- `Content-Type: application/json`
- `Accept: application/json`
- `Authorization: Bearer <access_token>` (for protected endpoints)
- `X-Request-ID`: Generated automatically or propagated across all responses.

### 1.3 Response Envelopes

#### Success Response
```json
{
  "success": true,
  "data": { ... },
  "meta": {
    "request_id": "a8b150afb8866b7c",
    "timestamp": "2026-08-22T07:15:00Z"
  }
}
```

#### Error Response
```json
{
  "success": false,
  "error": {
    "code": "INVALID_CREDENTIALS",
    "message": "invalid username or password",
    "request_id": "a8b150afb8866b7c",
    "timestamp": "2026-08-22T07:15:00Z",
    "details": null
  }
}
```

---

## 2. Authentication Endpoints (`/api/v1/auth`)

### 2.1 Register New Player
Creates a new student account with Argon2id password encryption and returns an access/refresh token pair.

- **URL**: `POST /api/v1/auth/register`
- **Auth Required**: No (Rate Limited: 30 req/min)
- **Request Body**:
  ```json
  {
    "username": "PencilLegend",
    "email": "legend@school.test",
    "password": "SecretPassword123!"
  }
  ```
- **Validation Rules**:
  - `username`: 3-30 chars, alphanumeric + dashes/underscores (`^[a-zA-Z0-9_-]{3,30}$`).
  - `email`: Optional valid RFC email address.
  - `password`: Minimum 6 characters.
- **Success Response (201 Created)**:
  ```json
  {
    "success": true,
    "data": {
      "tokens": {
        "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
        "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
        "expires_in": 900,
        "token_type": "Bearer"
      },
      "user": {
        "id": "usr_7x9p2a1b",
        "username": "PencilLegend",
        "email": "legend@school.test",
        "is_guest": false,
        "avatar_preset": "pencil_sketch_1",
        "title": "Classroom Rookie",
        "rating": 1200,
        "created_at": "2026-08-22T07:15:00Z",
        "updated_at": "2026-08-22T07:15:00Z"
      }
    }
  }
  ```
- **Error Codes**:
  - `400 BAD_REQUEST`: Validation failure.
  - `409 USER_EXISTS`: Username or email already registered.
  - `429 RATE_LIMIT_EXCEEDED`: Too many requests.

---

### 2.2 Login
Authenticates an existing player with their username/email and password.

- **URL**: `POST /api/v1/auth/login`
- **Auth Required**: No (Rate Limited: 30 req/min)
- **Request Body**:
  ```json
  {
    "username": "PencilLegend",
    "password": "SecretPassword123!"
  }
  ```
- **Success Response (200 OK)**:
  ```json
  {
    "success": true,
    "data": {
      "tokens": {
        "access_token": "eyJhbGci...",
        "refresh_token": "eyJhbGci...",
        "expires_in": 900,
        "token_type": "Bearer"
      },
      "user": {
        "id": "usr_7x9p2a1b",
        "username": "PencilLegend",
        "is_guest": false,
        "avatar_preset": "pencil_sketch_1",
        "title": "Classroom Rookie",
        "rating": 1200
      }
    }
  }
  ```
- **Error Codes**:
  - `401 INVALID_CREDENTIALS`: Password does not match or user not found.

---

### 2.3 Refresh Token (Token Rotation)
Exchanges an active refresh token for a brand new access token and a rotated refresh token. The previous refresh token is immediately revoked.

- **URL**: `POST /api/v1/auth/refresh`
- **Auth Required**: No
- **Request Body**:
  ```json
  {
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
  ```
- **Success Response (200 OK)**:
  ```json
  {
    "success": true,
    "data": {
      "tokens": {
        "access_token": "eyJhbGciOiJIUzI1Ni...",
        "refresh_token": "eyJhbGciOiJIUzI1Ni...",
        "expires_in": 900,
        "token_type": "Bearer"
      },
      "user": { ... }
    }
  }
  ```
- **Error Codes**:
  - `401 INVALID_REFRESH_TOKEN`: Token expired, malformed, or previously revoked.

---

### 2.4 Logout
Revokes the provided refresh token session.

- **URL**: `POST /api/v1/auth/logout`
- **Auth Required**: No (Optional Bearer token or payload)
- **Request Body**:
  ```json
  {
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
  ```
- **Success Response (200 OK)**:
  ```json
  {
    "success": true,
    "data": {
      "message": "logged out successfully"
    }
  }
  ```

---

### 2.5 Instant Guest Session
Mints a temporary guest student profile for instant play without registration.

- **URL**: `POST /api/v1/auth/guest`
- **Auth Required**: No
- **Request Body** *(Optional)*:
  ```json
  {
    "nickname": "DeskRacer",
    "avatar_preset": "pencil_sketch_guest"
  }
  ```
- **Success Response (201 Created)**: Returns tokens and guest player profile.

---

## 3. Users Endpoints (`/api/v1/users`)

### 3.1 Get Current Authenticated Profile
- **URL**: `GET /api/v1/users/me`
- **Auth Required**: **Yes (`Authorization: Bearer <access_token>`)**
- **Success Response (200 OK)**:
  ```json
  {
    "success": true,
    "data": {
      "user": {
        "id": "usr_7x9p2a1b",
        "username": "PencilLegend",
        "email": "legend@school.test",
        "is_guest": false,
        "avatar_preset": "pencil_sketch_1",
        "title": "Classroom Rookie",
        "rating": 1200
      },
      "games_played": 14,
      "games_won": 10,
      "win_rate": 71.4,
      "game_stats": [
        { "game_type": "hand_cricket", "rating": 1260, "played": 6, "won": 5, "lost": 1 },
        { "game_type": "dots_boxes", "rating": 1220, "played": 4, "won": 3, "lost": 1 }
      ]
    }
  }
  ```
- **Error Codes**:
  - `401 UNAUTHORIZED`: Missing or invalid Bearer token.

---

## 4. Summary of Auth Security Specifications

1. **Password Encryption**: Pure **Argon2id** (Memory: 64MB, Iterations: 3, Parallelism: 2, Salt: 16 cryptographically secure random bytes).
2. **Access Token TTL**: 15 minutes.
3. **Refresh Token TTL**: 7 days with active server-side rotation and instant logout revocation.
4. **Brute-Force Protection**: Sliding-window rate limiter on auth routes (30 requests/minute per client IP).
