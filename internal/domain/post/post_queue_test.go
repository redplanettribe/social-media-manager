package post

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPostQueue_Add(t *testing.T) {
	q := Queue{}
	q.Add("1")
	q.Add("2")
	q.Add("3")

	assert.Equal(t, 3, q.Len())
	assert.True(t, q.Contains("1"))
	assert.True(t, q.Contains("2"))
	assert.True(t, q.Contains("3"))
}

func TestPostQueue_Remove(t *testing.T) {
	q := Queue{}
	q.Add("1")
	q.Add("2")
	q.Add("3")

	q.Remove("2")

	assert.Equal(t, 2, q.Len())
	assert.True(t, q.Contains("1"))
	assert.False(t, q.Contains("2"))
	assert.True(t, q.Contains("3"))
}

func TestPostQueue_RemoveAt(t *testing.T) {
	q := Queue{}
	q.Add("1")
	q.Add("2")
	q.Add("3")

	q.RemoveAt(1)

	assert.Equal(t, 2, q.Len())
	assert.True(t, q.Contains("1"))
	assert.False(t, q.Contains("2"))
	assert.True(t, q.Contains("3"))
}

func TestPostQueue_InsertAt(t *testing.T) {
	q := Queue{}
	q.Add("1")
	q.Add("2")
	q.Add("3")

	q.InsertAt(1, "4")

	assert.Equal(t, 4, q.Len())
	assert.True(t, q.Contains("1"))
	assert.True(t, q.Contains("4"))
	assert.True(t, q.Contains("2"))
	assert.True(t, q.Contains("3"))
}

func TestPostQueue_Contains(t *testing.T) {
	q := Queue{}
	q.Add("1")
	q.Add("2")
	q.Add("3")

	assert.True(t, q.Contains("1"))
	assert.True(t, q.Contains("2"))
	assert.True(t, q.Contains("3"))
	assert.False(t, q.Contains("4"))
}

func TestPostQueue_IsEmpty(t *testing.T) {
	q := Queue{}
	assert.True(t, q.IsEmpty())

	q.Add("1")
	assert.False(t, q.IsEmpty())
}

func TestPostQueue_Len(t *testing.T) {
	q := Queue{}
	assert.Equal(t, 0, q.Len())

	q.Add("1")
	assert.Equal(t, 1, q.Len())

	q.Add("2")
	assert.Equal(t, 2, q.Len())
}

func TestPostQueue_Move(t *testing.T) {
	q := Queue{}
	q.Add("1")
	q.Add("2")
	q.Add("3")
	q.Add("4")

	q.Move(1, 2)

	assert.Equal(t, 4, q.Len())
	assert.True(t, q.Contains("1"))
	assert.True(t, q.Contains("3"))
	assert.True(t, q.Contains("2"))
	assert.True(t, q.Contains("4"))
	assert.Equal(t, "2", q.Get(2))
	assert.Equal(t, "3", q.Get(1))
}

func TestPostQueue_Get(t *testing.T) {
	q := Queue{}
	q.Add("1")
	q.Add("2")
	q.Add("3")

	assert.Equal(t, "1", q.Get(0))
	assert.Equal(t, "2", q.Get(1))
	assert.Equal(t, "3", q.Get(2))
}

func TestPostQueue_GetFirst(t *testing.T) {
	q := Queue{}
	q.Add("1")
	q.Add("2")
	q.Add("3")

	assert.Equal(t, "1", q.GetFirst())
}

func TestPostQueue_Shift(t *testing.T) {
	q := Queue{}
	q.Add("1")
	q.Add("2")
	q.Add("3")

	assert.Equal(t, 3, q.Len())
	assert.Equal(t, "1", q.Shift())
	assert.Equal(t, 2, q.Len())
	assert.Equal(t, "2", q.GetFirst())
}

func TestPostQueue_Pop(t *testing.T) {
	q := Queue{}
	q.Add("1")
	q.Add("2")
	q.Add("3")

	assert.Equal(t, 3, q.Len())
	assert.Equal(t, "2", q.Pop(1))
	assert.Equal(t, 2, q.Len())
	assert.Equal(t, "3", q.Get(1))
}