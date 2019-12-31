package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type dictResponse struct {
	name    string
	success bool
	defs    []struct {
		Fl       string
		Shortdef []string `json:"shortdef"`
	}
	alternatives []string
}

func divideLine(l string) []string {
	if len([]rune(l)) <= 80 {
		return []string{l}
	}

	return joinColumn(strings.Fields(l))
}

func joinColumn(s []string) []string {
	var ret []string
	var current string
	for _, word := range s {
		if len([]rune(current))+len([]rune(word))+1 > 80 {
			ret = append(ret, current)
			current = word + " "
			continue
		}
		current += word + " "
	}
	if current != "" {
		ret = append(ret, current)
	}
	return ret
}

func (dr *dictResponse) responseContent() string {
	if dr.success {
		defs := ""
		for i := range dr.defs {
			defs += dr.name + ": " + dr.defs[i].Fl + "\n"
			for j := range dr.defs[i].Shortdef {
				line := dr.defs[i].Shortdef[j]
				for _, l := range divideLine(line) {
					defs += "\t" + l + "\n"
				}
			}
		}
		return defs
	} else {
		return "Did you mean:\n\t" + strings.Join(dr.alternatives, "\n\t") + "\n"
	}
}

func dictQuery(apiKey, word string) (*dictResponse, error) {

	url := "https://dictionaryapi.com/api/v3/references/collegiate/json/" + word + "?key=" + apiKey
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("ERR: %#v\n", resp)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	dr := dictResponse{name: word}
	err = json.Unmarshal(body, &dr.defs)
	dr.success = true
	if err != nil {
		dr.success = false
		err = json.Unmarshal(body, &dr.alternatives)
		if err != nil {
			return nil, err
		}
	}
	return &dr, nil
}
