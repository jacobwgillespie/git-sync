package git

import "strings"

type Range struct {
	A string
	B string
}

func (r *Range) IsIdentical() bool {
	return strings.EqualFold(r.A, r.B)
}

func (r *Range) IsAncestor() bool {
	return Quiet("merge-base", "--is-ancestor", r.A, r.B)
}
