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

func TestFilter(t *testing.T) {
	RegisterTestingT(t)

	type givenDetail struct{}
	type whenDetail[T any] struct {
		slice []T          // 输入切片
		test  func(T) bool // 过滤条件函数
	}
	type thenExpected[T any] struct {
		result []T // 预期的过滤结果
	}

	tests := []testutil.Case[givenDetail, any, any]{
		{
			Scenario: "Filter empty slice",
			When:     "filtering an empty slice",
			Then:     "should return empty slice",
			WhenDetail: whenDetail[int]{
				slice: []int{},
				test:  func(v int) bool { return v > 0 },
			},
			ThenExpected: thenExpected[int]{
				result: []int{},
			},
		},
		{
			Scenario: "Filter int slice with partial matches",
			When:     "filtering int slice where some elements match even number condition",
			Then:     "should return only even numbers",
			WhenDetail: whenDetail[int]{
				slice: []int{1, 2, 3, 4, 5},
				test:  func(v int) bool { return v%2 == 0 },
			},
			ThenExpected: thenExpected[int]{
				result: []int{2, 4},
			},
		},
		{
			Scenario: "Filter string slice with prefix condition",
			When:     "filtering string slice where elements start with 'a'",
			Then:     "should return strings starting with 'a'",
			WhenDetail: whenDetail[string]{
				slice: []string{"apple", "banana", "avocado", "grape"},
				test:  func(v string) bool { return v[0] == 'a' },
			},
			ThenExpected: thenExpected[string]{
				result: []string{"apple", "avocado"},
			},
		},
		{
			Scenario: "Filter float32 slice with range condition",
			When:     "filtering float32 slice where elements between 2.0 and 4.0",
			Then:     "should return elements in range",
			WhenDetail: whenDetail[float32]{
				slice: []float32{1.5, 2.3, 3.7, 4.2, 5.0},
				test:  func(v float32) bool { return v > 2.0 && v < 4.0 },
			},
			ThenExpected: thenExpected[float32]{
				result: []float32{2.3, 3.7},
			},
		},
		{
			Scenario: "Filter with no matches",
			When:     "filtering slice where no elements meet condition",
			Then:     "should return empty slice",
			WhenDetail: whenDetail[int]{
				slice: []int{1, 3, 5},
				test:  func(v int) bool { return v%2 == 0 },
			},
			ThenExpected: thenExpected[int]{
				result: []int{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Scenario, func(t *testing.T) {
			// 类型断言转换泛型参数
			switch detail := tt.WhenDetail.(type) {
			case whenDetail[int]:
				res := Filter(detail.slice, detail.test)
				Expect(res).To(Equal(tt.ThenExpected.(thenExpected[int]).result))
			case whenDetail[string]:
				res := Filter(detail.slice, detail.test)
				Expect(res).To(Equal(tt.ThenExpected.(thenExpected[string]).result))
			case whenDetail[float32]:
				res := Filter(detail.slice, detail.test)
				Expect(res).To(Equal(tt.ThenExpected.(thenExpected[float32]).result))
			default:
				t.Fatalf("unsupported type in test case")
			}
		})
	}
}
