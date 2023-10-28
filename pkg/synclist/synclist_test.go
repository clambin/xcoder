package synclist_test

import (
	"github.com/clambin/vidconv/pkg/synclist"
	"github.com/stretchr/testify/assert"
	"slices"
	"testing"
)

func TestList_Add(t *testing.T) {
	s := synclist.UniqueList[string]{}

	assert.True(t, s.Add("foo"))
	assert.False(t, s.Add("foo"))
	assert.True(t, s.Add("bar"))
	assert.False(t, s.Add("bar"))

	s.Remove("bar")

	assert.False(t, s.Add("foo"))
	assert.True(t, s.Add("bar"))
}

func TestList_Remove(t *testing.T) {
	s := synclist.UniqueList[string]{}

	assert.False(t, s.Remove("foo"))
	s.Add("foo")
	s.Add("bar")
	assert.True(t, s.Remove("foo"))
	assert.False(t, s.Remove("foo"))
	assert.True(t, s.Remove("bar"))
	assert.False(t, s.Remove("bar"))
}

func TestList_List(t *testing.T) {
	s := synclist.UniqueList[string]{}

	assert.Empty(t, s.List())
	s.Add("foo")
	s.Add("bar")

	content := s.List()
	slices.Sort(content)
	assert.Equal(t, []string{"bar", "foo"}, content)
	assert.Equal(t, []string{"bar", "foo"}, s.ListOrdered())
}
