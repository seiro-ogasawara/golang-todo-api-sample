package usecase

import "fmt"

const (
	titleMaxLength       = 50
	descriptionMaxLength = 500
)

func validateTitle(title string) error {
	length := len(title)
	if length < 1 || length > titleMaxLength {
		return fmt.Errorf("length of title must be < %d, but %d", titleMaxLength, length)
	}
	return nil
}

func validateDescription(desc string) error {
	length := len(desc)
	if length > descriptionMaxLength {
		return fmt.Errorf("length of description must be < %d, but %d", descriptionMaxLength, length)
	}
	return nil
}
