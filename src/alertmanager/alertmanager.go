package alertmanager

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/parnurzeal/gorequest"
	"github.com/rymdo/kube-cron-rollout-restart/v2/src/types"
)

type Silence struct {
	Status struct {
		State string `json:"state"`
	} `json:"status"`
	ID        string    `json:"id"`
	Comment   string    `json:"comment"`
	CreatedBy string    `json:"createdBy"`
	UpdatedAt string    `json:"updatedAt"`
	EndsAt    string    `json:"endsAt"`
	StartsAt  string    `json:"startsAt"`
	Matchers  []Matcher `json:"matchers"`
}

type Matcher struct {
	IsRegex bool   `json:"isRegex"`
	Name    string `json:"name"`
	Value   string `json:"value"`
}

const (
	silencesURL = "/api/v1/silences"
	silenceURL  = "/api/v1/silence"
)

type Config struct {
	Url string
}

type Alertmanager struct {
	Config Config
}

func New(config Config) Alertmanager {
	return Alertmanager{}
}

func parseLabels(labelsStr string) map[string]string {
	results := make(map[string]string)

	labels := strings.Split(labelsStr, ",")
	for _, label := range labels {
		kv := strings.Split(label, "=")
		if len(kv) == 2 {
			results[kv[0]] = kv[1]
		}
	}

	return results
}

func generateSilence(silenceSource types.AlertmanagerSilence) Silence {
	comment := silenceSource.Comment
	if comment == "" {
		comment = "kube-cron-rollout-restart"
	}
	duration := silenceSource.Duration
	if duration <= 0 {
		duration = 15
	}

	silence := Silence{
		StartsAt:  time.Now().UTC().Format(time.RFC3339),
		EndsAt:    time.Now().Add(time.Minute * time.Duration(duration)).UTC().Format(time.RFC3339),
		Comment:   comment,
		CreatedBy: "kube-cron-rollout-restart",
		UpdatedAt: "0001-01-01T00:00:00Z",
	}

	for k, v := range parseLabels(silenceSource.Labels) {
		matcher := Matcher{
			IsRegex: false,
			Name:    k,
			Value:   v,
		}
		silence.Matchers = append(silence.Matchers, matcher)
	}

	return silence
}

func httpPost(url, data string, result chan<- error) {
	request := gorequest.New()
	resp, _, errs := request.Post(url).Send(data).End()

	if errs != nil {
		var errsStr []string
		for _, e := range errs {
			errsStr = append(errsStr, fmt.Sprintf("%s", e))
		}
		result <- fmt.Errorf("%s", strings.Join(errsStr, "; "))
		return
	}

	if resp.StatusCode != 200 {
		result <- fmt.Errorf("HTTP response code: %s", resp.Status)
		return
	}
	result <- nil
}

func (a *Alertmanager) CreateSilence(silenceSource types.AlertmanagerSilence) error {
	timeout := 3

	silence := generateSilence(silenceSource)

	fmt.Printf("Creating silence [creator: %s, comment: %s, start: %s, end: %s]\n", silence.CreatedBy, silence.Comment, silence.StartsAt, silence.EndsAt)

	json, err := json.Marshal(silence)
	if err != nil {
		return err
	}

	result := make(chan error)
	go httpPost(a.Config.Url+silencesURL, string(json), result)

	select {
	case err := <-result:
		return err
	case <-time.After(time.Second * time.Duration(timeout)):
		return errors.New("Alertmanager connections timeout")
	}
}
