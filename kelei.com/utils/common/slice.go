package common

func InsertStringSlice(slice, insertion []string, index int) []string {
	return append(slice[:index], append(insertion, slice[index:]...)...)
}

func InsertIntSlice(slice, insertion []int, index int) []int {
	return append(slice[:index], append(insertion, slice[index:]...)...)
}
