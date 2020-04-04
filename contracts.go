package sbproclient

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
)

type Contract struct {
	Month int
	Year  int
}

type Contracts []*Contract

func parseContract(b []byte) (*Contract, error) {
	fields := bytes.Split(b, []byte("-"))
	if l := len(fields); l != 2 {
		return nil, fmt.Errorf("got %d fields, expected 2", l)
	}

	month, err := strconv.Atoi(string(fields[0]))
	if err != nil {
		return nil, err
	}
	year, err := strconv.Atoi(string(fields[1]))
	if err != nil {
		return nil, err
	}

	if month < 1 && month > 12 {
		return nil, fmt.Errorf("wrong month number: %d", month)
	}

	if year >= 50 {
		year = year + 1900
	} else {
		year = year + 2000
	}

	contract := Contract{
		Month: month,
		Year:  year}

	return &contract, nil
}

func (contract *Contract) String() string {
	monthStr := fmt.Sprintf("%02d", contract.Month)
	yearStr := fmt.Sprintf("%d", contract.Year)

	return monthStr + "-" + yearStr[2:]
}

func (contracts *Contracts) SortDesc() {
	sort.Sort(sort.Reverse(contracts))
}

func (contracts Contracts) Len() int { return len(contracts) }

func (contracts Contracts) Less(i, j int) bool {
	if contracts[i].Year < contracts[j].Year {
		return true
	}
	if contracts[i].Year > contracts[j].Year {
		return false
	}

	return contracts[i].Month < contracts[j].Month
}

func (contracts Contracts) Swap(i, j int) {
	contracts[i], contracts[j] = contracts[j], contracts[i]
}
