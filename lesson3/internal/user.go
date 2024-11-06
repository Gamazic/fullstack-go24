package internal

type User struct {
	Id          *int
	Username    *string `json:"username"`
	Password    *string
	Email       *string
	PhoneNumber *string
	FirstName   *string
	LastName    *string
}
