# Subscription Service API Documentation

## Service Methods

### CreateSubscription
Creates a new subscription for a user to a specific laboratory work.

**Signature:**
```go
CreateSubscription(ctx context.Context, req *CreateSubscriptionReq) error
```

**Request Fields:**
- `UserUUID` (uuid.UUID): User's unique identifier
- `LabType` (LabType): Type of laboratory work (`Defence` or `Performance`)
- `LabTopic` (LabTopic): Topic of laboratory work (`Virtual`, `Electricity`, or `Mechanics`)
- `LabNumber` (int): Laboratory work number
- `LabAuditorium` (*int): Auditorium number (optional, depends on `LabType`)
- `CreatedAt` (time.Time): Creation timestamp

**Field Constraints:**
- `LabType`: Must be either `Defence` or `Performance`
- `LabTopic`: Must be one of `Virtual`, `Electricity`, or `Mechanics`
- `LabNumber`: Must be between 1 and 255 (inclusive)
- `LabAuditorium`:
    - Must be `nil` for `Defence` lab type (defence can happen in any auditorium)
    - Must NOT be `nil` for `Performance` lab type (performance requires specific auditorium)

**Errors:**
- `ValidationError`: One or more fields failed validation
- `ErrCreateSubscription`: Failed to create subscription in database

**Notes:**
- Subscription UUID is generated automatically by the service
- New subscriptions are created with `ClosedAt = nil` (active state)
- The combination of (LabType, LabTopic, LabNumber, LabAuditorium, UserUUID) forms a composite primary key

---

### CreateSubscriptionData
Creates subscription-related user data including time preferences, teacher blacklist, and statistics. This operation is transactional and must be called with an active transaction.

**Signature:**
```go
CreateSubscriptionData(ctx context.Context, tx pgx.Tx, req *CreateSubscriptionDataReq) error
```

**Request Fields:**
- `UserUUID` (uuid.UUID): User's unique identifier
- `TimePreferences` (map[DayOfWeek][]int): User's preferred time slots
    - Key: Day of week (`MON`, `TUE`, `WED`, `THU`, `FRI`, `SAT`, `SUN`)
    - Value: Array of lesson numbers (e.g., [1, 2, 3] means lessons 1, 2, and 3)
- `BlacklistedTeachers` ([]string): List of teacher names the user wants to avoid

**Field Constraints:**
- No validation is performed on `TimePreferences` or `BlacklistedTeachers`
- Empty arrays are allowed

**Errors:**
- `ErrCreateSubscription`: Failed to create subscription data in database (transaction will be rolled back)

**Notes:**
- This method commits the transaction upon successful completion
- Initial values are set: `SuccessfulSubscriptions = 0`, `LastSuccessfulSubscription = nil`
- Creates entries in three tables: `details`, `teacher_preferences`, and `time_preferences`
- The caller is responsible for opening the transaction and handling rollback on error

---

### GetSubscription
Retrieves a subscription by its UUID.

**Signature:**
```go
GetSubscription(ctx context.Context, subscriptionUUID uuid.UUID) (*GetSubscriptionRes, error)
```

**Parameters:**
- `subscriptionUUID` (uuid.UUID): Subscription's unique identifier

**Returns:**
- `GetSubscriptionRes` with fields:
    - `SubscriptionUUID` (uuid.UUID)
    - `LabType` (LabType)
    - `LabTopic` (LabTopic)
    - `LabNumber` (int)
    - `LabAuditorium` (*int)
    - `CreatedAt` (time.Time)
    - `ClosedAt` (*time.Time): `nil` if subscription is active

**Errors:**
- `ErrSubscriptionNotFound`: Subscription does not exist
- `error`: Database query failed

---

### GetSubscriptions
Retrieves all subscriptions for a specific user.

**Signature:**
```go
GetSubscriptions(ctx context.Context, userUUID uuid.UUID) ([]GetSubscriptionRes, error)
```

**Parameters:**
- `userUUID` (uuid.UUID): User's unique identifier

**Returns:**
- Array of `GetSubscriptionRes` (see `GetSubscription` for field details)
- Returns empty array if user has no subscriptions

**Errors:**
- `error`: Database query failed

**Notes:**
- Returns both active and closed subscriptions
- Results are not sorted by default

---

### UpdateSubscription
Updates an existing subscription's parameters.

**Signature:**
```go
UpdateSubscription(ctx context.Context, req *UpdateSubscriptionDataReq) error
```

**Request Fields:**
- `UserUUID` (uuid.UUID): User's unique identifier
- `SubscriptionUUID` (uuid.UUID): Subscription's unique identifier
- `LabType` (LabType): Type of laboratory work
- `LabTopic` (LabTopic): Topic of laboratory work
- `LabNumber` (int): Laboratory work number
- `LabAuditorium` (*int): Auditorium number (optional, depends on `LabType`)

**Field Constraints:**
Same as `CreateSubscription`

**Errors:**
- `ValidationError`: One or more fields failed validation
- `ErrUpdateSubscription`: Failed to update subscription in database

**Notes:**
- Cannot update `CreatedAt` or `ClosedAt` fields through this method
- Use `CloseSubscription` or `RestoreSubscription` to modify `ClosedAt`

---

### CloseSubscription
Closes an active subscription by setting `ClosedAt` to current timestamp.

**Signature:**
```go
CloseSubscription(ctx context.Context, subscriptionUUID uuid.UUID) error
```

**Parameters:**
- `subscriptionUUID` (uuid.UUID): Subscription's unique identifier

**Errors:**
- `ErrCloseSubscription`: Failed to close subscription in database

**Notes:**
- Sets `ClosedAt = NOW()` in the database
- Idempotent operation: closing an already closed subscription succeeds without error
- Closed subscriptions are excluded from matching in `GetMatchingSubscriptions`

---

### RestoreSubscription
Restores a closed subscription by setting `ClosedAt` to `nil`.

**Signature:**
```go
RestoreSubscription(ctx context.Context, subscriptionUUID uuid.UUID) error
```

**Parameters:**
- `subscriptionUUID` (uuid.UUID): Subscription's unique identifier

**Errors:**
- `ErrRestoreSubscription`: Failed to restore subscription in database

**Notes:**
- Sets `ClosedAt = nil` in the database
- Idempotent operation: restoring an already active subscription succeeds without error
- Restored subscriptions become eligible for matching in `GetMatchingSubscriptions`

---

### DeleteSubscription
Permanently deletes a subscription from the database.

**Signature:**
```go
DeleteSubscription(ctx context.Context, subscriptionUUID uuid.UUID) error
```

**Parameters:**
- `subscriptionUUID` (uuid.UUID): Subscription's unique identifier

**Errors:**
- `ErrDeleteSubscription`: Failed to delete subscription from database

**Notes:**
- This is a hard delete operation and cannot be undone
- Consider using `CloseSubscription` instead if you need to preserve historical data
- Idempotent operation: deleting a non-existent subscription succeeds without error

---

### GetMatchingSubscriptions
Finds all active subscriptions that match the given laboratory work parameters and available time slots, considering user preferences and teacher blacklists.

**Signature:**
```go
GetMatchingSubscriptions(ctx context.Context, req *GetMatchingSubscriptionsReq) ([]GetMatchingSubscriptionsRes, error)
```

**Request Fields:**
- `LabType` (LabType): Type of laboratory work
- `LabTopic` (LabTopic): Topic of laboratory work
- `LabNumber` (int): Laboratory work number
- `LabAuditorium` (int): Auditorium number
- `AvailableSlots` (map[DayOfWeek]map[int][]string): Available time slots with teachers
    - First key: Day of week
    - Second key: Lesson number
    - Value: Array of teacher names available for that slot

**Returns:**
- Array of `GetMatchingSubscriptionsRes` with fields:
    - `UserUUID` (uuid.UUID): Matching user's identifier
    - `SubscriptionUUID` (uuid.UUID): Matching subscription's identifier
    - `SuccessfulSubscriptions` (int): Number of successful subscriptions user has completed
    - `LastSuccessfulSubscription` (*time.Time): Date of user's last successful subscription
    - `MatchingTimeslots` (map[DayOfWeek][]int): Time slots that match user's preferences
        - Key: Day of week
        - Value: Array of lesson numbers that work for the user

**Matching Logic:**
A subscription matches if ALL of the following conditions are met:
1. `LabType`, `LabTopic`, `LabNumber`, and `LabAuditorium` exactly match the request
2. Subscription is active (`ClosedAt IS NULL`)
3. At least one time slot exists where:
    - The day and lesson are in the user's `TimePreferences`
    - At least one teacher in that slot is NOT in the user's `BlacklistedTeachers`

**Sorting:**
Results are sorted by:
1. `SuccessfulSubscriptions` (ascending) - users with fewer successful subscriptions first
2. `LastSuccessfulSubscription` (ascending, NULL values first) - users who haven't had recent success first

**Errors:**
- `error`: Database query failed

**Example:**

Request:
```go
GetMatchingSubscriptionsReq{
    LabType: "Defence",
    LabTopic: "Virtual",
    LabNumber: 3,
    LabAuditorium: 201,
    AvailableSlots: map[DayOfWeek]map[int][]string{
        "MON": {
            1: []string{"Ivanov", "Petrov"},
            2: []string{"Sidorov"},
        },
        "WED": {
            2: []string{"Ivanov"},
        },
    },
}
```

Given a user with:
- Time preferences: MON [1, 2, 3], WED [2]
- Blacklisted teachers: ["Ivanov"]

Result will include:
```go
MatchingTimeslots: map[DayOfWeek][]int{
    "MON": []int{1, 2}, // Lesson 1 has Petrov (not blacklisted), Lesson 2 has Sidorov
    "WED": []int{},     // Lesson 2 only has Ivanov (blacklisted), so excluded
}
```

Wait, that's wrong. Let me reconsider. If WED lesson 2 only has Ivanov who is blacklisted, then WED shouldn't appear in the result at all. Let me correct:

```go
MatchingTimeslots: map[DayOfWeek][]int{
    "MON": []int{1, 2}, // Lesson 1 has Petrov (not blacklisted), Lesson 2 has Sidorov
}
```

**Performance Notes:**
- Typical load: ~10 RPS
- Peak load (registration opening): up to 100 RPS expected
- Expected data volume: 25-100 subscriptions initially, potentially growing to hundreds
- Most subscriptions are filtered out by the lab parameter match before time slot matching

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
    "LabType": "must be either Defence or Performance",
    "LabNumber": "must be between 1 and 255",
    "LabAuditorium": "must not be nil for Performance lab type"
}
```

### Domain Errors
- `ErrSubscriptionNotFound`: Subscription does not exist in the database
- `ErrCreateSubscription`: Failed to create subscription or subscription-related data
- `ErrUpdateSubscription`: Failed to update subscription
- `ErrCloseSubscription`: Failed to close subscription
- `ErrRestoreSubscription`: Failed to restore subscription
- `ErrDeleteSubscription`: Failed to delete subscription

---

## Data Types

### LabType
```go
type LabType string

const (
    LabTypeDefence     LabType = "Defence"
    LabTypePerformance LabType = "Performance"
)
```

**Defence**: Laboratory work can be defended in any auditorium (auditorium assignment is flexible)
**Performance**: Laboratory work must be performed in a specific auditorium (auditorium assignment is fixed)

### LabTopic
```go
type LabTopic string

const (
    LabTopicVirtual     LabTopic = "Virtual"
    LabTopicElectricity LabTopic = "Electricity"
    LabTopicMechanics   LabTopic = "Mechanics"
)
```

### DayOfWeek
```go
type DayOfWeek string

const (
    DayMon DayOfWeek = "MON"
    DayTue DayOfWeek = "TUE"
    DayWed DayOfWeek = "WED"
    DayThu DayOfWeek = "THU"
    DayFri DayOfWeek = "FRI"
    DaySat DayOfWeek = "SAT"
    DaySun DayOfWeek = "SUN"
)
```