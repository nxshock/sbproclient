package sbproclient

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseContract(t *testing.T) {
	tests := []struct {
		s []byte
		e *Contract
	}{
		{s: []byte("03-13"), e: &Contract{Month: 3, Year: 2013}},
		{s: []byte("03-50"), e: &Contract{Month: 3, Year: 1950}},
		{s: []byte("03-51"), e: &Contract{Month: 3, Year: 1951}},
	}

	for _, test := range tests {
		gotContract, err := parseContract(test.s)
		assert.NoError(t, err)
		assert.Equal(t, test.e, gotContract)
	}
}

func TestContractString(t *testing.T) {
	contract := Contract{Month: 3, Year: 2020}
	assert.Equal(t, "03-20", contract.String())
}

func TestSortContracts(t *testing.T) {
	contracts := Contracts{
		&Contract{Month: 3, Year: 2000},
		&Contract{Month: 6, Year: 2000}}

	expectedContracts := Contracts{
		&Contract{Month: 6, Year: 2000},
		&Contract{Month: 3, Year: 2000}}

	contracts.SortDesc()

	assert.Equal(t, expectedContracts, contracts)
}
