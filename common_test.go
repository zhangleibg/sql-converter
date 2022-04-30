package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetGoStructType(t *testing.T) {
	inputs := []string{
		"bigint",
		"varchar",
		"aa",
	}
	expecteds := []string{
		"int64",
		"string",
		"interface{}",
	}

	for i, input := range inputs {
		runName := fmt.Sprintf("Case %d", i)
		t.Run(runName, func(t *testing.T) {
			assert.Equal(t, expecteds[i], mapping.getGoStructType(input).getString())
		})

	}

}
