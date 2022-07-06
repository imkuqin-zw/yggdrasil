package xarray

func ReverseStringArray(ss []string) {
	for i := len(ss)/2 - 1; i >= 0; i-- {
		opp := len(ss) - 1 - i
		ss[i], ss[opp] = ss[opp], ss[i]
	}
}

func RemoveReplaceStrings(arr []string) []string {
	set := make(map[string]struct{})
	j := 0
	for _, item := range arr {
		if _, ok := set[item]; ok {
			continue
		}
		arr[j] = item
		set[item] = struct{}{}
		j++
	}
	return arr[:j]
}
