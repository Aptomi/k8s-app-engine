package visibility

import "github.com/Frostman/aptomi/pkg/slinga"

type Entry map[string]interface{}

func GetServiceViewObject(state slinga.ServiceUsageState) interface{} {
	return Entry{
		"nodes": getNodes(),
		"edges": getEdges(),
	}
}

func getNodes() []Entry {
	result := []Entry{}
	result = append(result, Entry{"id": 0, "label": "API 0", "group": "source"})
	result = append(result, Entry{"id": 1, "label": "API 1", "group": "icons"})
	result = append(result, Entry{"id": 2, "label": "API 2", "group": "icons"})
	result = append(result, Entry{"id": 3, "label": "API 3", "group": "icons"})
	result = append(result, Entry{"id": 4, "label": "API 4", "group": "icons"})
	result = append(result, Entry{"id": 5, "label": "API 5", "group": "icons"})
	result = append(result, Entry{"id": 6, "label": "API 6", "group": "icons"})
	result = append(result, Entry{"id": 7, "label": "API 7", "group": "icons"})
	result = append(result, Entry{"id": 8, "label": "API 8", "group": "icons"})
	result = append(result, Entry{"id": 9, "label": "API 9", "group": "icons"})
	result = append(result, Entry{"id": 10, "label": "API 10", "group": "mints"})
	result = append(result, Entry{"id": 11, "label": "API 11", "group": "mints"})
	result = append(result, Entry{"id": 12, "label": "API 12", "group": "mints"})
	result = append(result, Entry{"id": 13, "label": "API 13", "group": "mints"})
	result = append(result, Entry{"id": 14, "label": "API 14", "group": "mints"})
	result = append(result, Entry{"id": 15, "group": "dotsWith"})
	result = append(result, Entry{"id": 16, "group": "dotsWith"})
	result = append(result, Entry{"id": 17, "group": "dotsWith"})
	result = append(result, Entry{"id": 18, "group": "dotsWith"})
	result = append(result, Entry{"id": 19, "group": "dotsWith"})
	result = append(result, Entry{"id": 20, "label": "API diamonds", "group": "diamonds"})
	result = append(result, Entry{"id": 21, "label": "API diamonds", "group": "diamonds"})
	result = append(result, Entry{"id": 22, "label": "API diamonds", "group": "diamonds"})
	result = append(result, Entry{"id": 23, "label": "API diamonds", "group": "diamonds"})
	return result
}

func getEdges() []Entry {
	result := []Entry{}
	result = append(result, Entry{"from": 1, "to": 0})
	result = append(result, Entry{"from": 2, "to": 0})
	result = append(result, Entry{"from": 4, "to": 3})
	result = append(result, Entry{"from": 5, "to": 4})
	result = append(result, Entry{"from": 4, "to": 0})
	result = append(result, Entry{"from": 7, "to": 6})
	result = append(result, Entry{"from": 8, "to": 7})
	result = append(result, Entry{"from": 7, "to": 0})
	result = append(result, Entry{"from": 10, "to": 9})
	result = append(result, Entry{"from": 11, "to": 10})
	result = append(result, Entry{"from": 10, "to": 4})
	result = append(result, Entry{"from": 13, "to": 12})
	result = append(result, Entry{"from": 14, "to": 13})
	result = append(result, Entry{"from": 13, "to": 0})
	result = append(result, Entry{"from": 16, "to": 15})
	result = append(result, Entry{"from": 17, "to": 15})
	result = append(result, Entry{"from": 15, "to": 10})
	result = append(result, Entry{"from": 19, "to": 18})
	result = append(result, Entry{"from": 20, "to": 19})
	result = append(result, Entry{"from": 19, "to": 4})
	result = append(result, Entry{"from": 22, "to": 21})
	result = append(result, Entry{"from": 23, "to": 22})
	result = append(result, Entry{"from": 23, "to": 0})
	return result
}