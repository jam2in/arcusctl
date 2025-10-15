package internal

type Task interface {
	Description() string
	Execute() error
}
