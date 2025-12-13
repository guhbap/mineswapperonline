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
	// MinComplexityRatio - минимальная сложность относительно референсного поля (25% от Dref)
	// Поля с меньшей сложностью не дают рейтинг, чтобы предотвратить фарм на очень простых полях
	MinComplexityRatio = 0.25
	// MinMines - минимальное количество мин для получения рейтинга
	// Предотвращает фарм на очень больших полях с малым количеством мин
	MinMines = 10
	// MinDensity - минимальная плотность мин (mines/area) для получения рейтинга
	// Референсное поле (16x16, 40) имеет плотность ~0.156 (15.6%)
	// Минимум установлен в 5% (0.05) для предотвращения фарма на больших полях с низкой плотностью
	MinDensity = 0.05
	// DFmin - минимальный коэффициент сложности поля (нижний порог для DF)
	DFmin = 0.05
	// Gamma - параметр для расчета DF (контроль крутизны)
	Gamma = 0.6
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

// ComputeComplexity computes complexity D for a field (exported version of computeD)
func ComputeComplexity(w, h, m float64) float64 {
	return computeD(w, h, m)
}

const α = 1.5

// CalculateDifficulty вычисляет сложность поля по формуле: (M / (W * H)) * sqrt(W^2 + H^2)
// где M - количество мин, W - ширина (cols), H - высота (rows)
func CalculateDifficulty(width, height, mines float64) float64 {
	cells := width * height
	density := mines / cells
	difficulty := cells * math.Pow(density, α)

	return difficulty
	// totalCells := width * height
	// if totalCells <= 0 {
	// 	return 0
	// }

	// density := mines / totalCells
	// diagonal := math.Sqrt(width*width + height*height)

	// return density * diagonal
}

// IsComplexitySufficient проверяет, достаточно ли сложности поля для получения рейтинга
// Возвращает true, если:
// 1. Количество мин >= MinMines
// 2. Плотность мин (mines/area) >= MinDensity
// 3. Сложность поля >= MinComplexityRatio * Dref
func IsComplexitySufficient(w, h, m float64, Dref float64) bool {
	if Dref <= 0 {
		return false
	}
	// Проверка минимального количества мин
	if m < MinMines {
		return false
	}
	// Проверка минимальной плотности мин
	A := w * h
	if A <= 0 {
		return false
	}
	density := m / A
	if density < MinDensity {
		return false
	}
	D := computeD(w, h, m)
	minComplexity := Dref * MinComplexityRatio
	return D >= minComplexity
}

// ComputeDifficultyFactor вычисляет коэффициент сложности поля DF
// DF = max(DFmin, (D/(D+Dref))^gamma)
// где D - сложность поля, Dref - референсная сложность, gamma - параметр крутизны
func ComputeDifficultyFactor(w, h, m, Dref float64) float64 {
	if Dref <= 0 {
		return DFmin
	}
	D := computeD(w, h, m)
	if D <= 0 {
		return DFmin
	}
	ratio := D / (D + Dref)
	df := math.Pow(ratio, Gamma)
	if df < DFmin {
		return DFmin
	}
	return df
}

// ComputeAttemptPoints вычисляет очки попытки P = K * DF * (S - E)
// где K - K-factor, DF - коэффициент сложности, S - оценка исполнения, E - ожидаемый результат
func ComputeAttemptPoints(w, h, m, T, Rpl, Dref float64) float64 {
	DF := ComputeDifficultyFactor(w, h, m, Dref)
	Texp := expectedTime(w, h, m)
	S := performanceScore(T, Texp)
	Rp := computeRp(w, h, m, Dref)
	E := expectedResult(Rp, Rpl)
	return K * DF * (S - E)
}

// UpdatePlayerRating performs one update, returns new rating and delta
// DEPRECATED: Используйте ComputeAttemptPoints и логику с BestP
func UpdatePlayerRating(w, h, m, T, Rpl float64, Dref float64) (newR float64, delta float64) {
	Rp := computeRp(w, h, m, Dref)
	Texp := expectedTime(w, h, m)
	S := performanceScore(T, Texp)
	E := expectedResult(Rp, Rpl)
	delta = K * (S - E)
	newR = Rpl + delta
	return
}
