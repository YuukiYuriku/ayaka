package sharedfunc

func UniqueStringSlice(input []string) []string {
	seen := make(map[string]struct{})
	var result []string
	for _, val := range input {
		if _, ok := seen[val]; !ok {
			seen[val] = struct{}{}
			result = append(result, val)
		}
	}
	return result
}