package command

type Command interface {
	Name() string
	Execute() error
}
