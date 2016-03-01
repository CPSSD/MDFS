package utils

func Prepend(arr []string, item string) []string {
	return append([]string{item}, arr...)
}