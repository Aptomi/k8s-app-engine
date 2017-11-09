package runtime

// AppendAll concatenates all provided info slices into a single info slice
func AppendAll(all ...[]*Info) []*Info {
	result := make([]*Info, 0)

	for _, infos := range all {
		result = append(result, infos...)
	}

	return result
}
