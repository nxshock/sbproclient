package sbproclient

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
)

type Symbol struct {
	Name      string
	TickCost  int // чтобы получить цену нужно поделить на это число
	Contracts []*Contract
}

func parseSymbols(b []byte) (symbols map[string]*Symbol, err error) {
	symbols = make(map[string]*Symbol)

	for _, symbolData := range bytes.Split(b, []byte("|")) {
		symbol, err := parseSymbol(symbolData)
		if err != nil {
			return nil, err
		}

		symbols[symbol.Name] = symbol
	}

	return symbols, nil
}

func parseSymbol(b []byte) (*Symbol, error) {
	fields := bytes.Split(b, []byte("!"))
	if l := len(fields); l != 2 {
		return nil, fmt.Errorf("expected 2 fields, got %d", l)
	}

	fields1 := bytes.Split(fields[0], []byte("_"))
	if l := len(fields1); l != 3 {
		return nil, fmt.Errorf("expected 3 fields, got %d", l)
	}

	tickCost, err := strconv.Atoi(string(fields1[2]))
	if err != nil {
		return nil, err
	}

	var contracts Contracts
	for _, constractCode := range bytes.Split(fields[1], []byte("*")) {
		contract, err := parseContract(constractCode)
		if err != nil {
			return nil, err
		}

		contracts = append(contracts, contract)
	}

	contracts.SortDesc()

	symbol := Symbol{
		Name:      string(fields1[0]),
		TickCost:  tickCost,
		Contracts: contracts}

	return &symbol, nil
}

func (symbol *Symbol) LatestContracts(n int) ([]*Contract, error) {
	if len(symbol.Contracts) == 0 {
		return nil, errors.New("no contracts available")
	}

	if len(symbol.Contracts) < n {
		return symbol.Contracts, nil
	}

	return symbol.Contracts[:n], nil
}
