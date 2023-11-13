package bdecoder

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

const (
	dictStart = 'd'
	dictEnd   = 'e'
	listStart = 'l'
	listEnd   = 'e'
	intStart  = 'i'
)

type DecodedObject map[string]interface{}

type bencodedInfo struct {
	Pieces      string
	PieceLength int
	Length      int
	Name        string
}

type BencodedTorrent struct {
	Announce string
	Info     bencodedInfo
}

func (bt BencodedTorrent) String() string {
	return fmt.Sprintf("Announce: %s\nInfo:\n  PieceLength: %d\n  Length: %d\n  Name: %s\n  Pieces: %s",
		bt.Announce, bt.Info.PieceLength, bt.Info.Length, bt.Info.Name, bt.Info.Pieces)
}

func decodeNumber(bencodedString string, i int) (int, int, error) {
	var decodedNumberStr string
	for i++; i < len(bencodedString) && bencodedString[i] != 'e'; i++ {
		decodedNumberStr += string(bencodedString[i])
	}
	decodedNumber, err := strconv.Atoi(decodedNumberStr)
	if err != nil {
		return 0, 0, fmt.Errorf("error decoding list: %v", err)
	}
	return i + 1, decodedNumber, nil
}

func decodeList(bencodedString string, i int) (int, []interface{}, error) {
	var items []interface{}
	var token interface{}
	var err error
	for i++; i < len(bencodedString) && bencodedString[i] != listEnd; {
		i, token, err = decodeToken(bencodedString, i)
		// If only one item in list, append the value alone and not a list containing the value
		if list, ok := token.([]interface{}); ok && len(list) == 1 {
			items = append(items, list[0])
		} else {
			items = append(items, token)
		}
	}
	if err != nil {
		return 0, items, fmt.Errorf("error decoding number: %v", err)
	}
	return i + 1, items, nil
}

func decodeToken(bencodedString string, i int) (int, interface{}, error) {
	var token interface{}
	var err error
	switch ch := bencodedString[i]; {
	case unicode.IsDigit(rune(ch)):
		i, token, err = decodeString(bencodedString, i)
	case ch == intStart:
		i, token, err = decodeNumber(bencodedString, i)
	case ch == listStart:
		i, token, err = decodeList(bencodedString, i)
	default:
		i, token, err = decodeObject(bencodedString, i)
	}
	if err != nil {
		return 0, token, err
	}
	return i, token, nil
}

func decodeString(bencodedString string, i int) (int, string, error) {
	var stringLenBuilder strings.Builder
	for ; unicode.IsNumber(rune(bencodedString[i])); i++ {
		stringLenBuilder.WriteRune(rune(bencodedString[i]))
	}
	stringLen, err := strconv.Atoi(stringLenBuilder.String())
	if err != nil {
		return 0, "", fmt.Errorf("error decoding string length: %v", err)
	}
	startIdx := i + 1
	endIdx := min(len(bencodedString)-1, i+stringLen)
	decodedString := bencodedString[startIdx : endIdx+1]
	return endIdx + 1, decodedString, nil
}

func decodeObject(bencodedString string, i int) (int, DecodedObject, error) {
	if i >= len(bencodedString) {
		return 0, nil, nil
	}
	if bencodedString[0] != dictStart || bencodedString[len(bencodedString)-1] != dictEnd {
		return 0, nil, errors.New("provided bencoded string has an invalid formatting")
	}
	decodedObj := make(DecodedObject)
	var err error
	var key string
	var token interface{}
	expectingKey := true
	for i++; i < len(bencodedString); {
		if expectingKey {
			i, key, err = decodeString(bencodedString, i)
			if key == "" {
				break // Might cause errors in the future. This is here because sometimes, the key returns empty. And by that time, the torrent data is actually fully decoded already (I think)
			}
			expectingKey = false
		} else {
			i, token, err = decodeToken(bencodedString, i)
			decodedObj[key] = token
			expectingKey = true
		}
		if err != nil {
			return 0, nil, err
		}
	}
	return i, decodedObj, nil
}

func Decode(bencodedString string) (BencodedTorrent, error) {
	bencodedTorrent := BencodedTorrent{}
	bencodedInfo := bencodedInfo{}
	_, decodedObj, err := decodeObject(bencodedString, 0)
	if err != nil {
		return bencodedTorrent, err
	}
	bencodedTorrent.Announce = decodedObj["announce"].(string)
	bencodedInfo.Pieces = decodedObj["info"].(DecodedObject)["pieces"].(string)
	bencodedInfo.PieceLength = decodedObj["info"].(DecodedObject)["piece length"].(int)
	bencodedInfo.Length = decodedObj["info"].(DecodedObject)["length"].(int)
	bencodedInfo.Name = decodedObj["info"].(DecodedObject)["name"].(string)
	bencodedTorrent.Info = bencodedInfo
	bencodedTorrent.Announce = decodedObj["announce"].(string)
	return bencodedTorrent, nil
}
