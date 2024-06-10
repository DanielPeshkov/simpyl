package algorithms

import (
	"math/rand"
	"simpyl/object"
	"testing"
	"time"
)

func TestCompareMergeAndQuickSort(t *testing.T) {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	array1 := make([]object.Object, random.Intn(100-10)+10)

	for i := range array1 {
		array1[i] = &object.Integer{Value: int64(random.Intn(100))}
	}
	array2 := make([]object.Object, len(array1))

	copy(array2, array1)
	low := 0
	high := len(array1) - 1
	QuickSort(array1, low, high)
	array2 = MergeSort(array2)

	for i := range array1 {
		if array1[i].(*object.Integer).Value != array2[i].(*object.Integer).Value {
			t.Fail()
		}
	}
}

func BenchmarkMergeSort(b *testing.B) {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	array := make([]object.Object, 1000000)
	for i := range array {
		array[i] = &object.Integer{Value: int64(random.Intn(100))}
	}
	b.ResetTimer()

	array = MergeSort(array)
}

func BenchmarkQuickSort(b *testing.B) {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	array := make([]object.Object, 1000000)
	for i := range array {
		array[i] = &object.Integer{Value: int64(random.Intn(100))}
	}
	b.ResetTimer()

	low := 0
	high := len(array) - 1
	QuickSort(array, low, high)
}
