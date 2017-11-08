package runtime

func AppendAll(all ...[]*Info) []*Info {
	result := make([]*Info, 0)

	for _, infos := range all {
		result = append(result, infos...)
	}

	return result
}
