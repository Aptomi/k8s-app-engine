package visibility

type lineEntry map[string]interface{}

type lineEntryList []lineEntry

func (list lineEntryList) Len() int {
	return len(list)
}

func (list lineEntryList) Less(i, j int) bool {
	return list[i]["id"].(string) < list[j]["id"].(string)
}

func (list lineEntryList) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
}
