package domain

type UserRepository interface {
	Create(user *User) error
	Get(id string) (*User, error)
	List() ([]User, error)
}
