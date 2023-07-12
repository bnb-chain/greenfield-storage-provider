package util

// ContainOnlyOneDifferentElement check two slices if they have the same length, order and one different element
func ContainOnlyOneDifferentElement(slice1, slice2 []uint32) int {
	if len(slice1) != len(slice2) {
		return -1
	}

	count := 0
	index := -1
	for i := 0; i < len(slice1); i++ {
		if slice1[i] != slice2[i] {
			count++
			index = i
		}
	}
	if count == 1 {
		return index
	}
	return -1
}
