package server

import (
	"math/rand"
	"time"
)

type Suggester interface {
	Suggest(tunnel *SshTunnel, attemptNum int) string
}

type UserNameSuggester struct{}

func (u UserNameSuggester) Suggest(tunnel *SshTunnel, attemptNum int) string {
	return tunnel.user
}

func NewUserNameSuggester() *UserNameSuggester {
	return &UserNameSuggester{}
}

type RandomizerSuggester struct {
	alphabet []rune
	length   int
}

func (r RandomizerSuggester) Suggest(tunnel *SshTunnel, attemptNum int) string {
	rand.Seed(time.Now().UnixNano())
	str := make([]rune, r.length)
	for i := range str {
		str[i] = r.alphabet[rand.Intn(len(r.alphabet))]
	}

	return string(str)
}

func NewRandomizerSuggester(length int) *RandomizerSuggester {
	return &RandomizerSuggester{
		alphabet: []rune("abcdefghijklmnopqrstuvwxyz"),
		length:   length,
	}
}

type CombinedSuggester struct {
	UserNameSuggester   *UserNameSuggester
	RandomizerSuggester *RandomizerSuggester
}

func (c CombinedSuggester) Suggest(tunnel *SshTunnel, attemptNum int) string {
	if attemptNum < 1 {
		return c.UserNameSuggester.Suggest(tunnel, attemptNum)
	}

	return c.RandomizerSuggester.Suggest(tunnel, attemptNum)
}

func NewCombinedSuggester() *CombinedSuggester {
	return &CombinedSuggester{
		UserNameSuggester:   NewUserNameSuggester(),
		RandomizerSuggester: NewRandomizerSuggester(3),
	}
}
