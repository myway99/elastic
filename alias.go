// Copyright 2012 Oliver Eilhard. All rights reserved.
// Use of this source code is governed by a MIT-license.
// See http://olivere.mit-license.org/license.txt for details.

package elastic

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type AliasService struct {
	client  *Client
	actions []aliasAction
	pretty  bool
	debug   bool
}

type aliasAction struct {
	// "add" or "remove"
	Type string
	// Index name
	Index string
	// Alias name
	Alias string
	// Filter
	Filter *Filter
}

func NewAliasService(client *Client) *AliasService {
	builder := &AliasService{
		client:  client,
		actions: make([]aliasAction, 0),
	}
	return builder
}

func (s *AliasService) Pretty(pretty bool) *AliasService {
	s.pretty = pretty
	return s
}

func (s *AliasService) Debug(debug bool) *AliasService {
	s.debug = debug
	return s
}

func (s *AliasService) Add(indexName string, aliasName string) *AliasService {
	action := aliasAction{Type: "add", Index: indexName, Alias: aliasName}
	s.actions = append(s.actions, action)
	return s
}

func (s *AliasService) AddWithFilter(indexName string, aliasName string, filter *Filter) *AliasService {
	action := aliasAction{Type: "add", Index: indexName, Alias: aliasName, Filter: filter}
	s.actions = append(s.actions, action)
	return s
}

func (s *AliasService) Remove(indexName string, aliasName string) *AliasService {
	action := aliasAction{Type: "remove", Index: indexName, Alias: aliasName}
	s.actions = append(s.actions, action)
	return s
}

func (s *AliasService) Do() (*AliasResult, error) {
	// Build url
	urls := "/_aliases"

	// Set up a new request
	req, err := s.client.NewRequest("POST", urls)
	if err != nil {
		return nil, err
	}

	// Parameters
	params := make(url.Values)
	if s.pretty {
		params.Set("pretty", fmt.Sprintf("%v", s.pretty))
	}
	urls += "?" + params.Encode()

	// Actions
	body := make(map[string]interface{})
	actionsJson := make([]interface{}, 0)

	for _, action := range s.actions {
		actionJson := make(map[string]interface{})
		detailsJson := make(map[string]interface{})
		detailsJson["index"] = action.Index
		detailsJson["alias"] = action.Alias
		if action.Filter != nil {
			detailsJson["filter"] = (*action.Filter).Source()
		}
		actionJson[action.Type] = detailsJson
		actionsJson = append(actionsJson, actionJson)
	}

	body["actions"] = actionsJson

	// Set body
	req.SetBodyJson(body)

	if s.debug {
		out, _ := httputil.DumpRequestOut((*http.Request)(req), true)
		fmt.Printf("%s\n", string(out))
	}

	// Get response
	res, err := s.client.c.Do((*http.Request)(req))
	if err != nil {
		return nil, err
	}
	if err := checkResponse(res); err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if s.debug {
		out, _ := httputil.DumpResponse(res, true)
		fmt.Printf("%s\n", string(out))
	}

	ret := new(AliasResult)
	if err := json.NewDecoder(res.Body).Decode(ret); err != nil {
		return nil, err
	}
	return ret, nil
}

// -- Result of an alias request.

type AliasResult struct {
	Ok  bool `json:"ok"`
	Ack bool `json:"acknowledged"`
}
