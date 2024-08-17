package router

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/metadata"
	"gopkg.in/yaml.v3"
)

type Endpoint struct {
	Name      string    `json:"name" yaml:"name"`
	Namespace string    `json:"namespace" yaml:"namespace"`
	IP        string    `json:"ip" yaml:"ip"`
	Port      string    `json:"port" yaml:"port"`
	Weight    int       `json:"weight" yaml:"weight"`
	Wants     []*Match  `json:"wants" yaml:"wants"`
	WantsType WantsType `json:"wants_type" yaml:"wants_type"`
}

type WantsType int

const (
	Wants_All_Match WantsType = iota
	Wants_Has_Match
	Wants_All_Not_Match
	Wants_Has_Not_Match
)

func (wt *WantsType) UnmarshalYAML(value *yaml.Node) error {
	switch strings.ToLower(value.Value) {
	case "all_match":
		*wt = Wants_All_Match
	case "has_match":
		*wt = Wants_Has_Match
	case "all_not_match":
		*wt = Wants_All_Not_Match
	case "has_not_match":
		*wt = Wants_Has_Not_Match
	default:
		return fmt.Errorf("wants_type must be one of all_match, has_match, all_not_match or has_not_match")
	}
	return nil
}

type Match struct {
	MatchType MatchType `json:"match_type" yaml:"match_type"`
	Pattern   string    `json:"pattern" yaml:"pattern"`
}

type MatchType int

const (
	Match_Prefix MatchType = iota
	Match_Precise
	Match_Regex
)

func (mt *MatchType) UnmarshalYAML(value *yaml.Node) error {
	switch strings.ToLower(value.Value) {
	case "prefix":
		*mt = Match_Prefix
	case "precise":
		*mt = Match_Precise
	case "regex":
		*mt = Match_Regex
	default:
		return fmt.Errorf("match_type must be one of prefix, precise or regex")
	}
	return nil
}

func (m *Match) isMatch(key string) bool {
	switch m.MatchType {
	case Match_Precise:
		return key == m.Pattern
	case Match_Prefix:
		return strings.HasPrefix(key, m.Pattern)
	case Match_Regex:
		match, err := regexp.MatchString(m.Pattern, key)
		return (err == nil) && match
	default:
		return false
	}
}

func (e *Endpoint) isAllMatch(key string) bool {
	for _, m := range e.Wants {
		if !m.isMatch(key) {
			return false
		}
	}
	return true
}

func (e *Endpoint) hasMatch(key string) bool {
	for _, m := range e.Wants {
		if m.isMatch(key) {
			return true
		}
	}
	return false
}

func (e *Endpoint) isAllNotMatch(key string) bool {
	for _, m := range e.Wants {
		if m.isMatch(key) {
			return false
		}
	}
	return true
}

func (e *Endpoint) hasNotMatch(key string) bool {
	for _, m := range e.Wants {
		if !m.isMatch(key) {
			return true
		}
	}
	return false
}

func (e *Endpoint) IsWants(key string) bool {
	switch e.WantsType {
	case Wants_All_Match:
		return e.isAllMatch(key)
	case Wants_Has_Match:
		return e.hasMatch(key)
	case Wants_All_Not_Match:
		return e.isAllNotMatch(key)
	case Wants_Has_Not_Match:
		return e.hasNotMatch(key)
	default:
		return false
	}
}

func (e *Endpoint) ToString() string {
	jsonData, err := json.Marshal(e)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	return string(jsonData)
}

func (e *Endpoint) ToAddr() string {
	return fmt.Sprintf("%v:%v", e.IP, e.Port)
}

func ParseEndpoint(s string) *Endpoint {
	var ep Endpoint
	err := json.Unmarshal([]byte(s), &ep)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	return &ep
}

func WantsToString(w []*Match) string {
	jsonData, err := json.Marshal(w)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	return string(jsonData)
}

func ParseWants(s string) []*Match {
	var w []*Match
	err := json.Unmarshal([]byte(s), &w)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	return w
}

type Metadata struct {
	Metadata map[string]string
	Ctx      context.Context
}

func (m *Metadata) Set(key, val string) {
	m.Metadata[key] = val
}

func (m *Metadata) SetCtx(ctx context.Context) {
	m.Ctx = ctx
}

func (m *Metadata) Get(key string) string {
	if val, exists := m.Metadata[key]; exists {
		return val
	}

	val := m.Ctx.Value(key)
	if val != nil {
		value, _ := val.(string)
		return value
	}

	return m.getFromGrpcMetadata(key)
}

func (m *Metadata) GetCtx() context.Context {
	return m.Ctx
}

func NewMetadata(ctx context.Context) *Metadata {
	return &Metadata{
		Metadata: map[string]string{},
		Ctx:      ctx,
	}
}

func (m *Metadata) getFromGrpcMetadata(key string) string {
	md, ok := metadata.FromOutgoingContext(m.Ctx)
	if !ok {
		return ""
	}
	keys := md.Get(key)
	if len(keys) == 0 {
		return ""
	}
	return keys[0]
}
