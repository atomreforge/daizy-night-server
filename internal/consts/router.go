package consts

type RouterDomain int

const (
	DomainNull         RouterDomain = iota // 0
	DomainUnauthorized                     // 1
	DomainPublic                           // 2
	DomainPrivate                          // 3
)

// Say renders the domain as its log expression.
func (d RouterDomain) Say() string {
	switch d {
	case DomainPublic:
		return string(DomainExprPublic)
	case DomainPrivate:
		return string(DomainExprPrivate)
	case DomainUnauthorized:
		return string(DomainExprUnauthorized)
	default:
		return string(DomainExprNull)
	}
}
