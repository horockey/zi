package main

import (
	"fmt"
	"math/big"
)

var (
	message = "Рассу"
	p       = 23
	q       = 61
)

func main() {
	fmt.Printf("Source message (M): \"%s\", p = %d, q = %d\n", message, p, q)
	n := p * q
	phi := (p - 1) * (q - 1)
	fmt.Printf("Phi = %d\n", phi)
	e := 2
	for ; e <= phi; e++ {
		if gcd(e, phi) == 1 {
			break
		}
	}
	if e > phi {
		panic("unable to find E")
	}

	d := 1
	for ; d < n; d++ {
		if (e*d)%phi == 1 {
			break
		}
	}
	if d >= n {
		panic("unable to find D")
	}

	fmt.Printf("Public key (E, N): %d, %d\n", e, n)
	fmt.Printf("Private key (D): %d\n", d)

	m := hash(message) % n
	fmt.Printf("Hashed message (m = h(M)): %d\n", m)

	M := new(big.Int)
	M.SetInt64(int64(m))

	N := new(big.Int)
	N.SetInt64(int64(n))

	sig := power(M, d)
	sig.Mod(sig, N)
	fmt.Printf("Signature (S) = %s\n", sig.String())

	m1 := power(sig, e)
	m1.Mod(m1, N)
	fmt.Printf("Recovered m (m`) = %d\n", m1)

	ciph := []int{}
	for _, c := range message {
		C := new(big.Int)
		C.SetInt64(int64(c))
		C.Mod(power(C, e), N)
		ciph = append(ciph, int(C.Int64()))
	}
	fmt.Printf("Cipher: %+v\n", ciph)

	recovered := []rune{}
	for _, c := range ciph {
		C := new(big.Int)
		C.SetInt64(int64(c))
		C.Mod(power(C, d), N)
		code := int(C.Int64())
		recovered = append(recovered, rune(code))
	}
	fmt.Printf("Recovered message: \"%s\"", string(recovered))
}

func gcd(a int, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	if a < 0 {
		return -a
	}
	return a
}

func power(base *big.Int, pwr int) *big.Int {
	res := new(big.Int)
	res.Set(base)

	for i := 1; i < pwr; i++ {
		res.Mul(res, base)
	}
	return res
}

func hash(message string) int {
	res := 0
	for _, c := range message {
		res += int(c)
	}
	return res
}
