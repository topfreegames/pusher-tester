package generators

type MessageGenerator interface {
	Generate() []byte
	Platform() string
}
