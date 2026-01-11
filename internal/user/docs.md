# User Service API Documentation

## Service Methods

### CreateUser
Creates a new user entity and returns its UUID.

**Signature:**
```go
CreateUser(ctx context.Context) (*CreateUserRes, error)
```

**Returns:**
- `CreateUserRes.UUID`: Generated UUID for the new user

**Errors:**
- `ErrCreateUser`: Failed to create user in database

---

### CreateUserDetails
Creates user details for an existing user.

**Signature:**
```go
CreateUserDetails(ctx context.Context, req *CreateUserDetailsReq) error
```

**Request Fields:**
- `UserUUID` (uuid.UUID): User's unique identifier
- `Name` (string): User's first name
- `Surname` (string): User's last name
- `Patronymic` (*string): User's patronymic (optional)
- `GroupCode` (string): Group identifier

**Field Constraints:**
- `Name`: Must contain only alphabetic characters, spaces, hyphens, underscores, and dots
- `Surname`: Must contain only alphabetic characters, spaces, hyphens, underscores, and dots
- `Patronymic`: If provided, must contain only alphabetic characters, spaces, hyphens, underscores, and dots
- `GroupCode`: Must match format `XX-YY-ZZ` (e.g., `AB-12-34`)
    - 2-3 alphabetic characters, followed by `-`, 1-2 digits, followed by `-`, 1-2 digits

**Errors:**
- `ValidationError`: One or more fields failed validation
- `ErrCreateUser`: Failed to create user details in database

---

### CreateUserContacts
Creates contact information for an existing user.

**Signature:**
```go
CreateUserContacts(ctx context.Context, req *CreateUserContactsReq) error
```

**Request Fields:**
- `UserUUID` (uuid.UUID): User's unique identifier
- `PhoneNumber` (string): User's phone number
- `Email` (*string): User's email address (optional)
- `TelegramID` (*int64): User's Telegram ID (optional)

**Field Constraints:**
- `PhoneNumber`: Must be in E.164 format (e.g., `+1234567890`)
    - Starts with `+`, followed by country code and number (1-15 total digits)
- `Email`: No validation currently applied
- `TelegramID`: If provided, must be a positive integer (> 0)

**Errors:**
- `ValidationError`: One or more fields failed validation
- `ErrCreateUser`: Failed to create user contacts in database

---

### GetUserInfo
Retrieves complete user information by UUID.

**Signature:**
```go
GetUserInfo(ctx context.Context, userUUID string) (*GetUserInfoRes, error)
```

**Parameters:**
- `userUUID` (string): User's UUID as string

**Returns:**
- `GetUserInfoRes` with fields:
    - `UUID` (uuid.UUID)
    - `Name` (string)
    - `Surname` (string)
    - `Patronymic` (*string)
    - `GroupCode` (string)
    - `PhoneNumber` (string)
    - `TelegramID` (*int64)

**Errors:**
- `error`: Invalid UUID format
- `ErrUserNotFound`: User does not exist
- `error`: Database query failed

---

### UpdateUserDetails
Updates user details for an existing user.

**Signature:**
```go
UpdateUserDetails(ctx context.Context, req *UpdateUserDetailsReq) error
```

**Request Fields:**
- `UserUUID` (uuid.UUID): User's unique identifier
- `Name` (string): User's first name
- `Surname` (string): User's last name
- `Patronymic` (*string): User's patronymic (optional)
- `GroupCode` (string): Group identifier

**Field Constraints:**
Same as `CreateUserDetails`

**Errors:**
- `ValidationError`: One or more fields failed validation
- `ErrUpdateUser`: Failed to update user details in database

---

### UpdateUserContacts
Updates contact information for an existing user.

**Signature:**
```go
UpdateUserContacts(ctx context.Context, req *UpdateUserContactsReq) error
```

**Request Fields:**
- `UserUUID` (uuid.UUID): User's unique identifier
- `PhoneNumber` (string): User's phone number
- `Email` (*string): User's email address (optional)
- `TelegramID` (*int64): User's Telegram ID (optional)

**Field Constraints:**
Same as `CreateUserContacts`

**Errors:**
- `ValidationError`: One or more fields failed validation
- `ErrUpdateUser`: Failed to update user contacts in database

---

## Error Types

### ValidationError
Contains a map of field names to error messages when validation fails.

```go
type ValidationError struct {
    Errors map[string]string
}
```

**Example:**
```json
{
    "Name": "must contain only alphabetic characters, spaces, hyphens, underscores, and dots",
    "GroupCode": "must match format XX-YY-ZZ (e.g., AB-12-34)"
}
```

### Domain Errors
- `ErrUserNotFound`: User does not exist in the database
- `ErrCreateUser`: Failed to create user or user-related data
- `ErrUpdateUser`: Failed to update user-related data