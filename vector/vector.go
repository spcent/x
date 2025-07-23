package vector

import (
	"math"
)

func Quantize(vec []float32) (quantized []int8, min, scale float32) {
	// Find the minimum and maximum values.
	min, max := float32(math.MaxFloat32), float32(-math.MaxFloat32)
	for _, v := range vec {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	// Calculate the quantization scale.
	scale = float32(255) / (max - min)

	// Quantize the data.
	quantized = make([]int8, len(vec))
	for i, v := range vec {
		quantized[i] = int8(math.Round(float64((v-min)*scale - 128)))
	}

	return quantized, min, scale
}

func Dequantize(quantized []int8, min, scale float32) []float32 {
	vec := make([]float32, len(quantized))
	for i, v := range quantized {
		vec[i] = (float32(v)+128)/scale + min
	}

	return vec
}
