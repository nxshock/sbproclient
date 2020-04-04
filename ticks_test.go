package sbproclient

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseTick(t *testing.T) {
	symbol, err := parseSymbol([]byte("6E_5_100000!03-16*03-17*03-18*03-19*03-20*06-16*06-17*06-18*06-19*06-20*09-16*09-17*09-18*09-19*12-16*12-17*12-18*12-19"))
	assert.NoError(t, err)

	tickData := "20190823 090503.003;110970;1;110965;110970;45501;10;15;0;0"

	loc, err := time.LoadLocation("America/Chicago")
	assert.NoError(t, err)

	client := &Client{ServerLocation: loc}

	expectedTick := Tick{
		Time:      time.Date(2019, 8, 23, 9, 5, 3, 3000000, loc),
		Cost:      1.1097,
		Volume:    1,
		Number:    45501,
		Direction: Ask,
		Pair:      symbol.Name,
		Contract:  symbol.Contracts[0].String()}

	gotTick, err := client.ParseTick(tickData, symbol.Name, symbol.Contracts[0].String(), symbol.TickCost)
	assert.NoError(t, err)
	assert.Equal(t, expectedTick, gotTick)
}

func TestParseResponse(t *testing.T) {
	s := "6E 03-20;6E 03-20:20200117 074545.991;111380;1;111380;111385;31412;11;66;0;0|20200117 074545.991;111380;2;111380;111385;31413;11;66;0;0|*"

	loc, err := time.LoadLocation("America/Chicago")
	assert.NoError(t, err)

	client := &Client{ServerLocation: loc}

	expected := []Tick{
		{
			Time:      time.Date(2020, 01, 17, 7, 45, 45, 991000000, loc),
			Cost:      1.1138,
			Volume:    1,
			Number:    31412,
			Direction: 1},
		{
			Time:      time.Date(2020, 01, 17, 7, 45, 45, 991000000, loc),
			Cost:      1.1138,
			Volume:    2,
			Number:    31413,
			Direction: 1},
	}

	response, maxTickNumber, err := client.ParseResponse(s, "", "", 100000)
	assert.NoError(t, err)
	assert.Equal(t, 31413, maxTickNumber)
	assert.Equal(t, expected, response)
}
