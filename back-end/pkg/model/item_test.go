package model

import (
	"testing"
)

func TestItemsElementIsEmptyReturnsFalse(t *testing.T) {
	input := ItemsElement{}

	if !input.IsEmpty() {
		t.Error("ItemsElement isEmpty should be true")
	}
}

func TestItemsElementContainsNeededFieldsReturnsTrue(t *testing.T) {
	input := ItemsElement{
		Name: "Name",
		Id:   "ID",
		Type: "Type",
	}
	if input.IsEmpty() {
		t.Error("ItemsElement isEmpty should be false")
	}
}
