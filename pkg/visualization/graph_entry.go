package visualization

type graphEntry map[string]interface{}

type graphEntryList []graphEntry

func (list graphEntryList) Len() int {
	return len(list)
}

func (list graphEntryList) Less(i, j int) bool {
	return list[i]["id"].(string) < list[j]["id"].(string)
}

func (list graphEntryList) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
}
