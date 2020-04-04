package sbproclient

import (
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func (client *Client) genFileName(date time.Time, symbolStr string, contractStr string) string {
	return fmt.Sprintf("cache/%s/%s/%s.zip", symbolStr, contractStr, date.In(client.ServerLocation).Format("20060102"))
}

func (client *Client) GetTicks(date time.Time, symbol *Symbol, contractStr string) ([]Tick, error) {
	client.logf("[%s, %s] загрузка тиков из истории\n", symbol.Name, contractStr)

	fileName := client.genFileName(date, symbol.Name, contractStr)

	if !fileExists(fileName) {
		client.logf("         файл отсутствует в кеше, загружаем с сервера\n")
		conn, err := net.Dial("tcp", defauktTicksHistoryServer.Address)
		if err != nil {
			return nil, err
		}
		defer conn.Close()

		request := []byte(fmt.Sprintf("<st>ticksx:%s:%s:%s</st>", symbol.Name, contractStr, date.Format("20060102")))

		_, err = conn.Write(request)
		if err != nil {
			return nil, err
		}

		sizeBytes := make([]byte, 1*1024)

		n, err := conn.Read(sizeBytes)
		if err != nil {
			return nil, err
		}

		sizeStr := string(sizeBytes[:n])
		sizeStr = strings.TrimPrefix(sizeStr, "<st>")
		sizeStr = strings.TrimSuffix(sizeStr, "</st>")

		if sizeStr == "no_file" {
			return nil, nil
		}

		nBytes, err := strconv.Atoi(sizeStr)
		if err != nil {
			return nil, err
		}

		b := make([]byte, nBytes)

		_, err = conn.Write([]byte("<st>ready</st>"))
		if err != nil {
			return nil, err
		}

		n, err = io.ReadFull(conn, b)
		if err != nil {
			return nil, err
		}

		if n == 0 {
			return nil, nil
		}

		os.MkdirAll(filepath.Dir(fileName), 0644)

		err = ioutil.WriteFile(fileName, b, 0644)
		if err != nil {
			return nil, err
		}
	}

	client.logf("         файл есть в кеше, считываем с диска\n")

	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	zipReader, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}

	var ticks []Tick

	buf := make([]byte, 26)

	i := 0
	for {
		n, err := io.ReadFull(zipReader, buf)
		if err == io.EOF {
			return ticks, nil
		} else if err != nil {
			return nil, err
		}

		if n < 26 {
			return ticks, fmt.Errorf("ожидалось считать 26 байт, получено %d", n)
		}

		i++

		tick, err := parseTick(date, buf, i, symbol, contractStr)
		if err != nil {
			return nil, err
		}

		ticks = append(ticks, tick)
	}

	client.logf("         загружено тиков из истории: %d\n", len(ticks))
	return ticks, nil
}

func parseTick(t time.Time, b []byte, n int, symbol *Symbol, contractStr string) (Tick, error) {
	secSinceZero := binary.LittleEndian.Uint32(b[0:4])
	cost := binary.LittleEndian.Uint32(b[8:12])
	volume := binary.LittleEndian.Uint32(b[12:16])

	var direction Direction
	switch {
	case b[16] == 0 || b[16] >= 128:
		direction = Bid
	case b[16] > 0 && b[16] < 128:
		direction = Ask
	default:
		return Tick{}, fmt.Errorf("unexpected tick direction: %v", b[16])
	}

	tick := Tick{
		Time:      t.Round(time.Hour * 24).Add(time.Duration(secSinceZero) * time.Second),
		Cost:      float64(cost) / float64(symbol.TickCost),
		Direction: direction,
		Number:    n,
		Volume:    int(volume),
		Pair:      symbol.Name,
		Contract:  contractStr}

	return tick, nil
}
