package models

type LazyAuth struct {
	key string
}

func NewLazyAuth() *LazyAuth {
	return &LazyAuth{
		key: "WelcomeToBadChess",
	}
}

func (a *LazyAuth) Authenticate(key string) bool {
	return key == a.key
}
