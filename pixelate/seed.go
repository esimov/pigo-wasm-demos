package pixelate

import (
	"image"
	"image/color"
	"math"
)

type prng struct {
	a    int
	m    int
	rand int
	div  float64
}

// addNoise applies a noise factor to the source image.
func addNoise(src *image.NRGBA64, amount int) {
	size := src.Bounds().Size()
	prng := &prng{
		a:    16807,
		m:    0x7fffffff,
		rand: 1.0,
		div:  1.0 / 0x7fffffff,
	}

	for x := 0; x < size.X; x++ {
		for y := 0; y < size.Y; y++ {
			noise := (prng.randomSeed() - 0.1) * float64(amount)
			r, g, b, a := src.At(x, y).RGBA()
			rf, gf, bf := float64(r), float64(g), float64(b)

			// Check if color do not overflow the maximum limit after noise has been applied
			if math.Abs(rf+noise) < 255 && math.Abs(gf+noise) < 255 && math.Abs(bf+noise) < 255 {
				rf += noise
				gf += noise
				bf += noise
			}
			r2 := max(0, min(255, uint8(rf)))
			g2 := max(0, min(255, uint8(gf)))
			b2 := max(0, min(255, uint8(bf)))

			src.Set(x, y, color.RGBA{R: r2, G: g2, B: b2, A: uint8(a)})
		}
	}
}

// nextLongRand generates a new random number based on the provided seed.
func (prng *prng) nextLongRand(seed int) int {
	lo := prng.a * (seed & 0xffff)
	hi := prng.a * (seed >> 16)
	lo += (hi & 0x7fff) << 16

	if lo > prng.m {
		lo &= prng.m
		lo++
	}
	lo += hi >> 15
	if lo > prng.m {
		lo &= prng.m
		lo++
	}
	return lo
}

// randomSeed returns a new random number.
func (prng *prng) randomSeed() float64 {
	prng.rand = prng.nextLongRand(prng.rand)
	return float64(prng.rand) * prng.div
}
