package vector

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/spcent/x/testutil"
)

func TestQuantizeDequantize(t *testing.T) {
	RegisterTestingT(t)

	type givenDetail struct{}
	type whenDetail struct {
		vector []float32
	}
	type thenExpected struct {
		maxError float32
	}

	tests := []testutil.Case[givenDetail, whenDetail, thenExpected]{
		{
			Scenario: "Quantize and dequantize unit vector",
			When:     "quantizing and then dequantizing a vector with values between 0 and 1",
			Then:     "should return vector close to the original with small error",
			WhenDetail: whenDetail{
				vector: []float32{0.1, 0.5, 0.9, 0.3},
			},
			ThenExpected: thenExpected{
				maxError: 0.01,
			},
		},
		{
			Scenario: "Quantize and dequantize vector with negative values",
			When:     "quantizing and then dequantizing a vector with negative values",
			Then:     "should return vector close to the original with small error",
			WhenDetail: whenDetail{
				vector: []float32{-1.0, -0.5, 0.0, 0.5, 1.0},
			},
			ThenExpected: thenExpected{
				maxError: 0.01,
			},
		},
		{
			Scenario: "Quantize and dequantize large range vector",
			When:     "quantizing and then dequantizing a vector with large range of values",
			Then:     "should return vector close to the original with acceptable error",
			WhenDetail: whenDetail{
				vector: []float32{-100, -50, 0, 50, 100},
			},
			ThenExpected: thenExpected{
				maxError: 1.5,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Scenario, func(t *testing.T) {
			// When.
			quantized, min, scale := Quantize(tt.WhenDetail.vector)
			dequantized := Dequantize(quantized, min, scale)

			// Then.
			Expect(len(dequantized)).To(Equal(len(tt.WhenDetail.vector)))

			maxError := float32(0)

			for i := range tt.WhenDetail.vector {
				error := float32(0)
				if tt.WhenDetail.vector[i] > dequantized[i] {
					error = tt.WhenDetail.vector[i] - dequantized[i]
				} else {
					error = dequantized[i] - tt.WhenDetail.vector[i]
				}
				if error > maxError {
					maxError = error
				}
			}

			Expect(maxError).To(BeNumerically("<=", tt.ThenExpected.maxError))
		})
	}
}
