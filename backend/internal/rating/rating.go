package rating

import (
	"math"
)

// Параметры, можно конфигурировать
const (
	Alpha = 0.7    // exponent for density in D
	Beta  = 0.5    // exponent in expected time
	C     = 4.0    // base time multiplier
	K     = 32.0   // K-factor for player rating updates
	Rref  = 1500.0 // rating for reference field (16x16,40)
	MinS  = 0.01
	MaxS  = 0.99
)

// computeD returns complexity D = A * (M/A)^alpha
func computeD(w, h, m float64) float64 {
	A := w * h
	if A <= 0 {
		return 0.0
	}
	density := m / A
	return A * math.Pow(density, Alpha)
}

// computeRp computes puzzle rating based on D, using Dref for ref field
func computeRp(w, h, m float64, Dref float64) float64 {
	D := computeD(w, h, m)
	if D <= 0 || Dref <= 0 {
		return Rref
	}
	return Rref + 400.0*math.Log10(D/Dref)
}

// expectedTime returns expected time in seconds for the field
func expectedTime(w, h, m float64) float64 {
	A := w * h
	if A <= 0 {
		return 1.0
	}
	density := m / A
	den := math.Pow(density, Beta)
	if den == 0 {
		den = 1e-9
	}
	return C * (math.Sqrt(A) / den)
}

// performanceScore S in (0.01..0.99) based on actual time T and expected time
func performanceScore(T, Texp float64) float64 {
	// linear mapping: S = 1 - (T - Texp) / (3*Texp)
	s := 1.0 - (T-Texp)/(3.0*Texp)
	if s < MinS {
		s = MinS
	}
	if s > MaxS {
		s = MaxS
	}
	return s
}

// expectedResult E by Elo formula
func expectedResult(Rp, Rpl float64) float64 {
	return 1.0 / (1.0 + math.Pow(10.0, (Rp-Rpl)/400.0))
}

// ComputeDref computes reference complexity D for reference field (16x16, 40 mines)
func ComputeDref() float64 {
	return computeD(16.0, 16.0, 40.0)
}

// UpdatePlayerRating performs one update, returns new rating and delta
func UpdatePlayerRating(w, h, m, T, Rpl float64, Dref float64) (newR float64, delta float64) {
	Rp := computeRp(w, h, m, Dref)
	Texp := expectedTime(w, h, m)
	S := performanceScore(T, Texp)
	E := expectedResult(Rp, Rpl)
	delta = K * (S - E)
	newR = Rpl + delta
	return
}
