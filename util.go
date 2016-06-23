package gmvc

import (
    "math"
)

func Round(f float64, n int) float64 {
    base := math.Pow10(n)

    f = f * base

    if f < 0 {
        f = math.Ceil(f - 0.5) / base
    } else {
        f = math.Floor(f + 0.5) / base
    }

    return f
}

