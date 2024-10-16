package main

type Priv uint8

const (
	privRead  Priv = 1 << 2
	privWrite Priv = 1 << 1
	privGrant Priv = 1
)

func (p Priv) canRead() bool {
	return (p>>2)%2 == 1
}

func (p Priv) canWrite() bool {
	return (p>>1)%2 == 1
}

func (p Priv) canGrant() bool {
	return p%2 == 1
}

func (p Priv) String() string {
	res := ""
	if p.canRead() {
		res += "R"
	}
	if p.canWrite() {
		res += "W"
	}
	if p.canGrant() {
		res += "G"
	}

	return res
}
