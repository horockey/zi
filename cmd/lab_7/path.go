package main

type Path []Coord

func (p Path) Matches(other Path) bool {
	if len(p) != len(other) {
		return false
	}
	for idx := range p {
		if p[idx] != other[idx] {
			return false
		}
	}

	return true
}
