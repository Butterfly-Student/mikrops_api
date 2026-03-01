package model

// User is a backward-compatibility alias for AdminUser.
// The old `users` table has been replaced by `admin_users`.
// Deprecated: use AdminUser instead.
type User = AdminUser

// UserInput is a backward-compatibility alias for AdminUserInput.
// Deprecated: use AdminUserInput instead.
type UserInput = AdminUserInput

// UserFilter is a backward-compatibility alias for AdminUserFilter.
// Deprecated: use AdminUserFilter instead.
type UserFilter = AdminUserFilter
