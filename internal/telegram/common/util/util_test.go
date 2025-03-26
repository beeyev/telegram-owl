package util_test

import (
	"testing"

	"github.com/beeyev/telegram-owl/internal/telegram/common/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStructToFormPayload_Positive(t *testing.T) {
	type TestStruct struct {
		Name     string
		Age      int
		IsActive bool
		Verified bool
		Balance  float64
	}

	input := TestStruct{
		Name:     "Alice",
		Age:      30,
		IsActive: true,
		Verified: false,
		Balance:  123.45,
	}

	result, err := util.StructToFormPayload(input)
	require.NoError(t, err)

	expected := map[string]string{
		"Name":     "Alice",
		"Age":      "30",
		"IsActive": "1",
		"Verified": "0",
		"Balance":  "123.45",
	}

	assert.Equal(t, expected, result)
}

func TestStructToFormPayload_NilInput(t *testing.T) {
	result, err := util.StructToFormPayload(nil)
	assert.Nil(t, result)
	assert.EqualError(t, err, "input struct is nil")
}

func TestStructToFormPayload_MarshalError(t *testing.T) {
	type Unmarshallable struct {
		Ch chan int // JSON can't marshal channels
	}
	input := Unmarshallable{Ch: make(chan int)}
	result, err := util.StructToFormPayload(input)
	assert.Nil(t, result)
	assert.ErrorContains(t, err, "failed to marshal struct")
}
