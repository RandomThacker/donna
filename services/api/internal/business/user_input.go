package business

// CreateUserInput is the domain input for registering a user.
type CreateUserInput struct {
	Email       string
	DisplayName *string
	AvatarURL   *string
	Timezone    string
	Locale      *string
}

// UpdateUserInput is a partial update. Nil pointers mean "leave unchanged".
type UpdateUserInput struct {
	DisplayName *string
	AvatarURL   *string
	Timezone    *string
	Locale      *string
	Status      *string
}
