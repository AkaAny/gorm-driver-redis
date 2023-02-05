package gorm_driver_redis

import "sort"

// 如何找出字典序升序的所有排序
// 从右向左找到第一个非递减(递减是指从左向右)元素的位置:descIndex,非递增元素:descElem
// 从右向左找到第一个大于descElem的元素位置:greaterIndex,元素:greaterElem
// 交换descElem和greaterElem
// descIndex后的元素按照升序排列
// 循环出口：元素降序排列
func intSliceNextPermutation(ori []int) (hasNext bool, next []int) {
	elemNum := len(ori)
	// 从右向左寻找第一组非递减的元素
	descIndex := -1
	for rIndex := elemNum - 1; rIndex > 0; rIndex-- {
		if ori[rIndex-1] <= ori[rIndex] {
			descIndex = rIndex - 1
			break
		}
	}
	if descIndex == -1 {
		return false, nil
	}
	// 从右向左寻找第一个大于ori[descIndex]的元素
	var greaterIndex int
	for rIndex := elemNum - 1; rIndex > descIndex; rIndex-- {
		if ori[rIndex] > ori[descIndex] {
			greaterIndex = rIndex
			break
		}
	}
	// 交换元素
	ori[descIndex], ori[greaterIndex] = ori[greaterIndex], ori[descIndex]
	// index + 1升序排列
	sort.SliceStable(ori[descIndex+1:], func(i, j int) bool {
		return ori[descIndex+1+i] < ori[descIndex+1+j]
	})
	next = make([]int, len(ori))
	copy(next, ori)
	return true, next
}

func sliceArrayCount(count int) int {
	if count == 0 {
		return 1
	}
	var result = 1
	for ; count > 0; count-- {
		result *= count
	}
	return result
}

// 升序排列->降序排列及过程中的所有排列构成元素的全排列
func GenIntSlicePermutationWithDict(ori []int) [][]int {
	resultSlice := make([][]int, 0, sliceArrayCount(len(ori)))
	// 升序排列
	sort.SliceStable(ori, func(i, j int) bool {
		return ori[i] < ori[j]
	})
	// 首个排列
	copyElem := make([]int, len(ori))
	copy(copyElem, ori)
	resultSlice = append(resultSlice, copyElem)
	for {
		hasNext, nextPermutation := intSliceNextPermutation(ori)
		if !hasNext {
			return resultSlice
		}
		resultSlice = append(resultSlice, nextPermutation)

	}
}
