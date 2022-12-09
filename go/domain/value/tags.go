package value

import "fmt"

type Tag struct {
	Id  int    `json:"id"`
	Tag string `json:"tag"`
}

// TODO (taekyeom) move to seperate data files
var TagMap = map[int]string{
	1: "반려동물",
	2: "빠른이동",
	3: "대화자제",
	4: "3~4명",
	5: "경유지",
	6: "큰짐",
}

func GetTagById(tagId int) (string, error) {
	tag, ok := TagMap[tagId]
	if !ok {
		return "", fmt.Errorf("%w: Unknown tag id %d", ErrInvalidOperation, tagId)
	}

	return tag, nil
}
