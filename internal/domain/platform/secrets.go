package platform

type Secrets string

func (s Secrets) String() string {
	return string(s)
}