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
}

func (i Items) IsEmpty() bool {
	return len(i.ItemElements) == 0
}

func (ie ItemsElement) IsEmpty() bool {
	//TODO: Maybe change empty check to only include name - create a
	// sep one for the below scenario
	//TODO: Is this going to cause an issue with the watchlist?
	return ie.Name == "" || ie.Id == "" || ie.Type == ""
}

func (ie ItemsElement) IsOfCorrectType(expectedType string) bool {
	return ie.Type == expectedType
}

func (ie ItemsElement) IsSameByYearAndName(matchingType ItemsElement) bool {
	return ie.Name == matchingType.Name && ie.ProductionYear == matchingType.ProductionYear
}

func (i Items) GetItemByName(name string) ItemsElement {
	for _, item := range i.ItemElements {
		for item.Name == name {
			return item
		}
	}
	return ItemsElement{}
}
