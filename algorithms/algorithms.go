package algorithms

import (
	"simpyl/object"
)

var typeMap = map[object.ObjectType]int{
	"INTEGER": 1,
	"FLOAT":   2,
	"STRING":  3,
}

func mergeCompare(a object.Object, b object.Object) bool {
	ind := 10*typeMap[a.Type()] + typeMap[b.Type()]

	switch ind {
	case 11:
		return a.(*object.Integer).Value <= b.(*object.Integer).Value
	case 12:
		return float64(a.(*object.Integer).Value) <= b.(*object.Float).Value
	case 13:
		return true

	case 21:
		return a.(*object.Float).Value <= float64(b.(*object.Integer).Value)
	case 22:
		return a.(*object.Float).Value <= b.(*object.Float).Value
	case 23:
		return true

	case 31:
		return false
	case 32:
		return false
	case 33:
		return a.(*object.String).Value <= b.(*object.String).Value

	default:
		panic("Invalid type comparison")
	}
}

func MergeSort(arr []object.Object) []object.Object {
	if len(arr) <= 1 {
		return arr
	}

	// find middle index
	middle := len(arr) / 2

	// divide array in half
	left := MergeSort(arr[:middle])
	right := MergeSort(arr[middle:])

	return merge(left, right)
}

func merge(left []object.Object, right []object.Object) []object.Object {
	result := make([]object.Object, 0, len(left)+len(right))

	for len(left) > 0 || len(right) > 0 {
		if len(left) == 0 {
			return append(result, right...)
		}

		if len(right) == 0 {
			return append(result, left...)
		}

		if mergeCompare(left[0], right[0]) {
			result = append(result, left[0])
			left = left[1:]
		} else {
			result = append(result, right[0])
			right = right[1:]
		}
	}

	return result
}

func quickCompare(a object.Object, b object.Object) bool {
	ind := 10*typeMap[a.Type()] + typeMap[b.Type()]

	switch ind {
	case 11:
		return a.(*object.Integer).Value < b.(*object.Integer).Value
	case 12:
		return float64(a.(*object.Integer).Value) < b.(*object.Float).Value
	case 13:
		return true

	case 21:
		return a.(*object.Float).Value < float64(b.(*object.Integer).Value)
	case 22:
		return a.(*object.Float).Value < b.(*object.Float).Value
	case 23:
		return true

	case 31:
		return false
	case 32:
		return false
	case 33:
		return a.(*object.String).Value < b.(*object.String).Value

	default:
		panic("Invalid type comparison")
	}
}

func QuickSort(arr []object.Object, low int, high int) {
	if low < high {
		pi := partition(arr, low, high)

		QuickSort(arr, low, pi-1)
		QuickSort(arr, pi+1, high)
	}
}

func partition(arr []object.Object, low int, high int) int {
	pivot := arr[high]
	i := low - 1

	for j := low; j < high; j++ {
		if quickCompare(arr[j], pivot) {
			i++
			arr[i], arr[j] = arr[j], arr[i]
		}
	}

	arr[i+1], arr[high] = arr[high], arr[i+1]
	return i + 1
}
