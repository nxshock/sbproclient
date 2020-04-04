package sbproclient

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"sync"
	"time"
)

type Client struct {
	Symbols        map[string]*Symbol
	ServerLocation *time.Location

	key string

	LogWriters []io.Writer
}

func NewClient(key string) (*Client, error) {
	b, err := gzipRequest(defaultSymbolsServer.Address, requestSymbols)
	if err != nil {
		return nil, err
	}

	symbols, err := parseSymbols(b)
	if err != nil {
		return nil, err
	}

	client := &Client{
		Symbols:        symbols,
		ServerLocation: defaultTicksServer.Location,
		key:            key}

	return client, nil
}

func (client *Client) request(conn net.Conn, scanner *bufio.Scanner, request string) (string, error) {
	requestBytes := append(append(begin, request...), end...)

	// client.logln(">", string(requestBytes))

	_, err := conn.Write(requestBytes)
	if err != nil {
		return "", err
	}

	if !scanner.Scan() {
		return "", fmt.Errorf("ошибка при сканировании ответа: %v", scanner.Err())
	}

	// client.logln("<", scanner.Text())

	return scanner.Text(), nil
}

func scanCmd(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if bytes.Index(data, begin) == -1 || bytes.Index(data, end) == -1 {
		if atEOF {
			return 0, nil, io.ErrUnexpectedEOF
		}

		return 0, nil, nil
	}

	i := bytes.Index(data, begin) + len(begin)
	j := bytes.Index(data, end)

	if j > i {
		return j + len(end), data[i:j], nil
	}

	return 0, nil, nil
}

func (client *Client) RequestTicks(pair, contract string, lastKnownTickNum, div int, conn net.Conn, scanner *bufio.Scanner) ([]Tick, int, error) {
	request := fmt.Sprintf("ticks:%s %s;%s %s;%d|", pair, contract, pair, contract, lastKnownTickNum)
	response, err := client.request(conn, scanner, request)
	if err != nil {
		return nil, 0, err
	}

	return client.ParseResponse(response, pair, contract, div)
}

func (client *Client) log(v ...interface{}) {
	if client.LogWriters == nil {
		return
	}

	for _, w := range client.LogWriters {
		fmt.Fprint(w, v...)
	}
}

func (client *Client) logln(v ...interface{}) {
	if client.LogWriters == nil {
		return
	}

	for _, w := range client.LogWriters {
		fmt.Fprintln(w, v...)
	}
}

func (client *Client) logf(format string, a ...interface{}) {
	if client.LogWriters == nil {
		return
	}

	for _, w := range client.LogWriters {
		fmt.Fprintf(w, format, a...)
	}
}

func gzipRequest(serverAddr string, requestBytes []byte) ([]byte, error) {
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	w := gzip.NewWriter(conn)

	_, err = w.Write(bytes.Join([][]byte{messagePrefix, requestBytes, messageSuffix}, nil))
	if err != nil {
		return nil, fmt.Errorf("write req error: %v", err)
	}

	err = w.Close()
	if err != nil {
		return nil, fmt.Errorf("close req error: %v", err)
	}

	r, err := gzip.NewReader(conn)
	if err != nil {
		return nil, fmt.Errorf("init reader error: %v", err)
	}

	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read resp error: %v", err)
	}

	b = bytes.TrimPrefix(b, []byte("<ms>"))
	b = bytes.TrimSuffix(b, []byte("</me>"))

	return b, nil
}

func (client *Client) OnlineUpdate(symbolStr string) (chan Tick, error) {
	symbol, exists := client.Symbols[symbolStr]
	if !exists {
		return nil, fmt.Errorf("symbol %s is not available", symbol.Name)
	}

	if symbol.TickCost <= 0 {
		return nil, fmt.Errorf("tickcost %d <= 0", symbol.TickCost)
	}

	contracts, err := symbol.LatestContracts(2) // TODO: переместить константу в управляемое место
	if err != nil {
		return nil, err
	}

	client.logf("Получение тиков для контрактов: %v\n", contracts)

	ticksChan := make(chan Tick)

	wg := new(sync.WaitGroup)
	wg.Add(len(contracts))

	go func() {
		defer close(ticksChan)

		wg.Wait()
	}()

	for _, contract := range contracts {
		go func(c *Contract) {
			for {
				conn, err := net.Dial("tcp", defaultTicksServer.Address)
				if err != nil {
					client.logf("update error: %v\n", err)
					return // TODO: handle error
				}
				defer conn.Close()

				s := bufio.NewScanner(conn)
				s.Split(scanCmd)

				client.request(conn, s, client.key)

				var lastKnownTickNum int
				for {
					var ticks []Tick
					ticks, newLastKnownTickNum, err := client.RequestTicks(symbolStr, c.String(), lastKnownTickNum, symbol.TickCost, conn, s)
					// client.logf("получено %d тиков, последний номер %d\n", len(ticks), lastKnownTickNum)
					if err != nil {
						client.logf("update error: %v\n", err)
						break // TODO: handle error
					}

					for _, tick := range ticks {
						ticksChan <- tick
					}

					if len(ticks) > 0 {
						lastKnownTickNum = newLastKnownTickNum
					}

					time.Sleep(2 * time.Second)
				}
				time.Sleep(10 * time.Second)
			}

			wg.Done()
		}(contract)
	}

	return ticksChan, nil
}

func (client *Client) LoadFromHistory() (ticks []Tick, err error) {
	//lastTickNums := make([]int, len(client.Symbols))

	startTime := time.Now().In(client.ServerLocation)

	for symbolStr, symbol := range client.Symbols {
		client.logf("Загрузка данных из истории %s за %s...", symbolStr, startTime.Format("02.01.2006"))

		latestContracts, err := symbol.LatestContracts(2)
		if err != nil {
			return nil, fmt.Errorf("ошибка при расчёте последних контрактов для %s: %v", symbolStr, err)
		}

		for _, contract := range latestContracts {
			ticks, err := client.GetTicks(startTime, symbol, contract.String())
			if err != nil {
				return nil, fmt.Errorf("ошибка при получении архива тиков с сервера: %v", err)
			}

			if len(ticks) == 0 {
				client.logf("Данных нет на сервере для %s.", symbolStr)
			}

			for _, tick := range ticks {
				ticks = append(ticks, tick)
			}
		}
	}

	return ticks, nil
}

func (client *Client) Close() {
	for _, v := range client.LogWriters {
		if c, e := v.(io.Closer); e {
			err := c.Close()
			if err != nil {
				log.Println("ошибка при закрытии лога:", err)
			}
		}
	}
}
