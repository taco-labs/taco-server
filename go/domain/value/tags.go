package value

import "fmt"

type Tag struct {
	Id  int    `json:"id"`
	Tag string `json:"tag"`
}

// TODO (taekyeom) move to seperate data files
var TagMap = map[int]string{
	1: "애견 동반",
}

func GetTagById(tagId int) (string, error) {
	tag, ok := TagMap[tagId]
	if !ok {
		return "", fmt.Errorf("%w: Unknown tag id %d", ErrInvalidOperation, tagId)
	}

	return tag, nil
}
