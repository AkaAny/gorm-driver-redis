package utils

func Contains[T comparable](u []T, sub T) bool {
	for _, item := range u {
		if item == sub {
			return true
		}
	}
	return false
}

type PredicateFunc[T any] func(item T) bool

func Filter[T any](u []T, predicateFunc PredicateFunc[T]) []T {
	var results = make([]T, 0)
	for _, item := range u {
		if !predicateFunc(item) {
			continue
		}
		results = append(results, item)
	}
	return results
}

func CrossCollection[T comparable](a []T, b []T) []T {
	var results = make([]T, 0)
	for _, aItem := range a {
		if Contains[T](b, aItem) {
			results = append(results, aItem)
		}
	}
	return results
}

type MapFunc[I any, O any] func(item I) O

func Map[I any, O any](u []I, mapFunc MapFunc[I, O]) []O {
	var results = make([]O, 0)
	for _, inputItem := range u {
		var outputItem = mapFunc(inputItem)
		results = append(results, outputItem)
	}
	return results
}

type IEquals interface {
	Equals(other interface{}) bool
}

func SliceEqual[T IEquals](a []T, b []T) bool {
	for _, aItem := range a {
		for _, bItem := range b {
			if !aItem.Equals(bItem) {
				return false
			}
		}
	}
	return true
}
