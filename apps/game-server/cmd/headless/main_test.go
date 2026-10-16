package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestProtocolRejectsInvalidInputAndRunsAnActualEpisode(t *testing.T) {
	input := `{"op":"step"}
{"op":"reset","seed":7,"agents":1,"opponents":0,"horizon":1}
{"op":"step","policy":"heuristic"}
{"op":"step","policy":"random"}
{"op":"reset","seed":8,"agents":5}
{"op":"reset","script":"arbitrary"}
{"op":"reset"} {"op":"step"}
{"op":"close"}
`
	var out bytes.Buffer
	serve(strings.NewReader(input), &out)
	decoder := json.NewDecoder(&out)
	var replies []response
	for decoder.More() {
		var r response
		if err := decoder.Decode(&r); err != nil {
			t.Fatal(err)
		}
		replies = append(replies, r)
	}
	if len(replies) != 8 || replies[0].Error == "" || replies[1].Result == nil || !replies[2].Result.Done || replies[2].Result.Hash == "" || !replies[7].Closed {
		t.Fatal(replies)
	}
	for _, i := range []int{3, 4, 5, 6} {
		if replies[i].Error == "" {
			t.Fatal("accepted invalid request", i)
		}
	}
}
