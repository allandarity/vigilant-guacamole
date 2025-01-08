package model

type Items struct {
	ItemElements []ItemsElement `json:"Items"`
}

type ItemsElement struct {
	Name            string  `json:"Name"`
	Id              string  `json:"Id"`
	Type            string  `json:"Type"`
	ProductionYear  int16   `json:"ProductionYear"`
	CommunityRating float32 `json:"CommunityRating"`
	Image           MovieImage
}

func (ie ItemsElement) IsEmpty() bool {
	return ie.Name == "" || ie.Id == "" || ie.Type == ""
}

func (ie ItemsElement) IsOfCorrectType(expectedType string) bool {
	return ie.Type == expectedType
}

func (i Items) GetItemByName(name string) ItemsElement {
	for _, item := range i.ItemElements {
		for item.Name == name {
			return item
		}
	}
	return ItemsElement{}
}
