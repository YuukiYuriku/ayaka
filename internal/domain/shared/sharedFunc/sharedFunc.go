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

func UniqueTupleSlice(pairs [][2]string) [][2]string {
	seen := make(map[string]bool)
	result := make([][2]string, 0)

	for _, p := range pairs {
		key := p[0] + "|" + p[1]
		if !seen[key] {
			result = append(result, p)
			seen[key] = true
		}
	}
	return result
}