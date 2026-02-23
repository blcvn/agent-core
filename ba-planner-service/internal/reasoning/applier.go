package chat_agent

import (
	"context"

	v32 "github.com/blcvn/ba-shared-libs/pkg/domain/v3.2"
)

// ModificationApplier applies modifications to the document
type ModificationApplier struct {
}

// NewModificationApplier creates a new ModificationApplier
func NewModificationApplier() *ModificationApplier {
	return &ModificationApplier{}
}

// ApplyModifications applies modifications to the document
func (a *ModificationApplier) ApplyModifications(ctx context.Context, doc *v32.Document, modifications []Modification) (*v32.Document, error) {
	// Create a copy... or modify in place?
	// Assuming doc is already a copy or safe to modify
	updatedDoc := doc

	for _, mod := range modifications {
		switch mod.ActionType {
		case "add":
			if err := a.applyAdd(updatedDoc, mod); err != nil {
				return nil, err
			}
		case "modify":
			if err := a.applyModify(updatedDoc, mod); err != nil {
				return nil, err
			}
		case "delete":
			if err := a.applyDelete(updatedDoc, mod); err != nil {
				return nil, err
			}
		}
	}

	return updatedDoc, nil
}

func (a *ModificationApplier) applyAdd(doc *v32.Document, mod Modification) error {
	// Logic depends on doc tier and section type
	// Need to parse new_content into struct and append
	return nil
}

func (a *ModificationApplier) applyModify(doc *v32.Document, mod Modification) error {
	// Logic depends on doc tier and section type
	// Find item by ID and replace
	return nil
}

func (a *ModificationApplier) applyDelete(doc *v32.Document, mod Modification) error {
	// Logic depends on doc tier and section type
	// Filter out item by ID
	return nil
}
