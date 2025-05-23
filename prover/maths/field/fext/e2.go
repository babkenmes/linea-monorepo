package fext

import (
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
)

const ExtensionDegree int = 2

// E2 is a degree two finite field extension of fr.Element
type E2 struct {
	A0, A1 fr.Element
}

// Equal returns true if z equals x, false otherwise
func (z *E2) Equal(x *E2) bool {
	return z.A0.Equal(&x.A0) && z.A1.Equal(&x.A1)
}

// Cmp compares (lexicographic order) z and x and returns:
//
//	-1 if z <  x
//	 0 if z == x
//	+1 if z >  x
func (z *E2) Cmp(x *E2) int {
	if a1 := z.A1.Cmp(&x.A1); a1 != 0 {
		return a1
	}
	return z.A0.Cmp(&x.A0)
}

// LexicographicallyLargest returns true if this element is strictly lexicographically
// larger than its negation, false otherwise
func (z *E2) LexicographicallyLargest() bool {
	// adapted from github.com/zkcrypto/bls12_381
	if z.A1.IsZero() {
		return z.A0.LexicographicallyLargest()
	}
	return z.A1.LexicographicallyLargest()
}

// SetString sets a E2 element from strings
func (z *E2) SetString(s1, s2 string) *E2 {
	z.A0.SetString(s1)
	z.A1.SetString(s2)
	return z
}

// SetZero sets an E2 elmt to zero
func (z *E2) SetZero() *E2 {
	z.A0.SetZero()
	z.A1.SetZero()
	return z
}

// Set sets an E2 from x
func (z *E2) Set(x *E2) *E2 {
	z.A0 = x.A0
	z.A1 = x.A1
	return z
}

// SetOne sets z to 1 in Montgomery form and returns z
func (z *E2) SetOne() *E2 {
	z.A0.SetOne()
	z.A1.SetZero()
	return z
}

// SetRandom sets a0 and a1 to random values
func (z *E2) SetRandom() (*E2, error) {
	if _, err := z.A0.SetRandom(); err != nil {
		return nil, err
	}
	if _, err := z.A1.SetRandom(); err != nil {
		return nil, err
	}
	return z, nil
}

// IsZero returns true if z is zero, false otherwise
func (z *E2) IsZero() bool {
	return z.A0.IsZero() && z.A1.IsZero()
}

// IsOne returns true if z is one, false otherwise
func (z *E2) IsOne() bool {
	return z.A0.IsOne() && z.A1.IsZero()
}

// Add adds two elements of E2
func (z *E2) Add(x, y *E2) *E2 {
	z.A0.Add(&x.A0, &y.A0)
	z.A1.Add(&x.A1, &y.A1)
	return z
}

// Sub subtracts two elements of E2
func (z *E2) Sub(x, y *E2) *E2 {
	z.A0.Sub(&x.A0, &y.A0)
	z.A1.Sub(&x.A1, &y.A1)
	return z
}

// Double doubles an E2 element
func (z *E2) Double(x *E2) *E2 {
	z.A0.Double(&x.A0)
	z.A1.Double(&x.A1)
	return z
}

// Neg negates an E2 element
func (z *E2) Neg(x *E2) *E2 {
	z.A0.Neg(&x.A0)
	z.A1.Neg(&x.A1)
	return z
}

// String implements Stringer interface for fancy printing
func (z *E2) String() string {
	return z.A0.String() + "+" + z.A1.String() + "*u"
}

// MulByElement multiplies an element in E2 by an element in fr
func (z *E2) MulByElement(x *E2, y *fr.Element) *E2 {
	var yCopy fr.Element
	yCopy.Set(y)
	z.A0.Mul(&x.A0, &yCopy)
	z.A1.Mul(&x.A1, &yCopy)
	return z
}

// Conjugate conjugates an element in E2
func (z *E2) Conjugate(x *E2) *E2 {
	z.A0 = x.A0
	z.A1.Neg(&x.A1)
	return z
}

// Halve sets z to z / 2
func (z *E2) Halve() {
	z.A0.Halve()
	z.A1.Halve()
}

// Legendre returns the Legendre symbol of z
func (z *E2) Legendre() int {
	var n fr.Element
	z.norm(&n)
	return n.Legendre()
}

// Exp sets z=xᵏ (mod q²) and returns it
func (z *E2) Exp(x E2, k *big.Int) *E2 {
	if k.IsUint64() && k.Uint64() == 0 {
		return z.SetOne()
	}

	e := k
	if k.Sign() == -1 {
		// negative k, we invert
		// if k < 0: xᵏ (mod q²) == (x⁻¹)ᵏ (mod q²)
		x.Inverse(&x)

		// we negate k in a temp big.Int since
		// Int.Bit(_) of k and -k is different
		e = bigIntPool.Get().(*big.Int)
		defer bigIntPool.Put(e)
		e.Neg(k)
	}

	z.SetOne()
	b := e.Bytes()
	for i := 0; i < len(b); i++ {
		w := b[i]
		for j := 0; j < 8; j++ {
			z.Square(z)
			if (w & (0b10000000 >> j)) != 0 {
				z.Mul(z, &x)
			}
		}
	}

	return z
}

// Sqrt sets z to the square root of and returns z
// The function does not test whether the square root
// exists or not, it's up to the caller to call
// Legendre beforehand.
// cf https://eprint.iacr.org/2012/685.pdf (algo 10)
func (z *E2) Sqrt(x *E2) *E2 {

	// precomputation
	var b, c, d, e, f, x0 E2
	var _b, o fr.Element

	// c must be a non square (works for p=1 mod 12 hence 1 mod 4, only bls377 has such a p currently)
	c.A1.SetOne()

	q := fr.Modulus()
	var exp, one big.Int
	one.SetUint64(1)
	exp.Set(q).Sub(&exp, &one).Rsh(&exp, 1)
	d.Exp(c, &exp)
	e.Mul(&d, &c).Inverse(&e)
	f.Mul(&d, &c).Square(&f)

	// computation
	exp.Rsh(&exp, 1)
	b.Exp(*x, &exp)
	b.norm(&_b)
	o.SetOne()
	if _b.Equal(&o) {
		x0.Square(&b).Mul(&x0, x)
		_b.Set(&x0.A0).Sqrt(&_b)
		z.Conjugate(&b).MulByElement(z, &_b)
		return z
	}
	x0.Square(&b).Mul(&x0, x).Mul(&x0, &f)
	_b.Set(&x0.A0).Sqrt(&_b)
	z.Conjugate(&b).MulByElement(z, &_b).Mul(z, &e)

	return z
}

// BatchInvertE2 returns a new slice with every element in a inverted.
// It uses Montgomery batch inversion trick.
//
// if a[i] == 0, returns result[i] = a[i]
func BatchInvertE2(a []E2) []E2 {
	res := make([]E2, len(a))
	if len(a) == 0 {
		return res
	}

	zeroes := make([]bool, len(a))
	var accumulator E2
	accumulator.SetOne()

	for i := 0; i < len(a); i++ {
		if a[i].IsZero() {
			zeroes[i] = true
			continue
		}
		res[i].Set(&accumulator)
		accumulator.Mul(&accumulator, &a[i])
	}

	accumulator.Inverse(&accumulator)

	for i := len(a) - 1; i >= 0; i-- {
		if zeroes[i] {
			continue
		}
		res[i].Mul(&res[i], &accumulator)
		accumulator.Mul(&accumulator, &a[i])
	}

	return res
}

// Select is conditional move.
// If cond = 0, it sets z to caseZ and returns it. otherwise caseNz.
func (z *E2) Select(cond int, caseZ *E2, caseNz *E2) *E2 {
	//Might be able to save a nanosecond or two by an aggregate implementation

	z.A0.Select(cond, &caseZ.A0, &caseNz.A0)
	z.A1.Select(cond, &caseZ.A1, &caseNz.A1)

	return z
}

// Div divides an element in E2 by an element in E2
func (z *E2) Div(x *E2, y *E2) *E2 {
	var r E2
	r.Inverse(y).Mul(x, &r)
	return z.Set(&r)
}
