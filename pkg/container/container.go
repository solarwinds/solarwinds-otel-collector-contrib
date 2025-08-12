package container

type Provider interface {
	ReadContainerInstanceID() (string, error)
	IsRunInContainerd() bool
}

func NewProvider() Provider {
	return newProvider()
}
