package lister

import (
	"testing"

	"github.com/KarelKubat/goto-meet/item"
)

func TestFirstNext(t *testing.T) {
	lister := &Lister{
		list: &List{
			Items: []*item.Item{
				{
					Title: "1",
				},
				{
					Title: "2",
				},
				{
					Title: "3",
				},
				{
					Title: "4",
				},
			},
		},
	}
	it := lister.First()
	if it.Title != "1" {
		t.Errorf("First() returns title %q, want 1", it.Title)
	}
	it = lister.Next()
	if it.Title != "2" {
		t.Errorf("Next() returns title %q, want 2", it.Title)
	}
	it = lister.Next()
	if it.Title != "3" {
		t.Errorf("Next() returns title %q, want 3", it.Title)
	}
	it = lister.Next()
	if it.Title != "4" {
		t.Errorf("Next() returns title %q, want 4", it.Title)
	}
	it = lister.Next()
	if it != nil {
		t.Errorf("Next() returns %v, want nil", it)
	}
}
