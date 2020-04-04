package sbproclient

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Direction int

const (
	Ask Direction = iota
	Bid
)

type Tick struct {
	// Время
	Time time.Time

	//Цена
	Cost float64

	// Количество
	Volume int

	// Номер тика
	Number int

	// Направление
	Direction Direction

	// Валютная пара
	Pair string

	// Контракт
	Contract string
}

func (client *Client) ParseTick(s, pair, contract string, div int) (Tick, error) {
	items := strings.Split(s, ";")
	if len(items) != 10 {
		return Tick{}, fmt.Errorf("unexpected item count: %d", len(items))
	}

	datetime, err := time.ParseInLocation("20060102 150405.000", items[0], client.ServerLocation)
	if err != nil {
		return Tick{}, err
	}

	costInt1, err := strconv.ParseInt(items[1], 10, 32)
	if err != nil {
		return Tick{}, err
	}
	cost1 := float64(costInt1) / float64(div)

	costInt2, err := strconv.ParseInt(items[3], 10, 32)
	if err != nil {
		return Tick{}, err
	}
	cost2 := float64(costInt2) / float64(div)

	costInt3, err := strconv.ParseInt(items[4], 10, 32)
	if err != nil {
		return Tick{}, err
	}
	cost3 := float64(costInt3) / float64(div)

	var direction Direction
	if cost2 < cost1 {
		direction = Ask
	} else if cost3 > cost1 {
		direction = Bid
	}

	volume, err := strconv.Atoi(items[2])
	if err != nil {
		return Tick{}, err
	}

	number, err := strconv.Atoi(items[5])
	if err != nil {
		return Tick{}, err
	}

	tick := Tick{
		Time:      datetime,
		Cost:      cost1,
		Volume:    volume,
		Number:    number,
		Direction: direction,
		Pair:      pair,
		Contract:  contract}

	return tick, nil
}

func (client *Client) ParseResponse(s, pair, contract string, div int) (response []Tick, maxTickNumber int, err error) {
	switch s {
	case "killapp":
		return nil, 0, ErrKillApp
	}

	fields := strings.Split(s, ":")
	if len(fields) != 2 {
		return nil, 0, fmt.Errorf("неожиданное количество полей: ожидалось два поля, найдено %d (%s)", len(fields), s)
	}
	ticksBytes := strings.Split(fields[1], "|")
	for _, tickBytes := range ticksBytes {
		if tickBytes == "*" {
			continue
		}

		tick, err := client.ParseTick(tickBytes, pair, contract, div)
		if err != nil {
			return nil, 0, err
		}
		response = append(response, tick)
		if tick.Number > maxTickNumber {
			maxTickNumber = tick.Number
		}
	}

	return response, maxTickNumber, nil
}
