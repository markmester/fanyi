/*
 * File: detect.go
 * Project: bot
 * File Created: Thursday, 26th January 2023 11:41:18 am
 * Author: Mark Mester (mmester6016@gmail.com)
 * -----
 * Last Modified: Saturday, 28th January 2023 4:15:49 pm
 * Modified By: Mark Mester (mmester6016@gmail.com>)
 */
package clients

import (
	"encoding/json"
	"errors"
	"log"
	"sort"
	"strings"

	"github.com/pemistahl/lingua-go"
)

const DefaultDetectionThreshold = 0.5

var commonLanguages = []lingua.Language{
	lingua.English,
	lingua.Chinese,
	lingua.German,
	lingua.Spanish,
	lingua.Japanese,
	lingua.French,
}

type Channel string
type Detector struct {
	linguaAllLanguages    lingua.LanguageDetector `json:"-"`
	linguaCommonLanguages lingua.LanguageDetector `json:"-"`

	SelectDetectors map[Channel]*SelectDetector `json:"select_detectors"` // -> maps channel to select detector
}

type SelectDetector struct {
	linguaSelectLanguages lingua.LanguageDetector `json:"-"`
	Selected              *Selected               `json:"selected"`
}

type Selected struct {
	L1 lingua.Language `json:"l1"`
	L2 lingua.Language `json:"l2"`
}

func NewDetector() *Detector {
	return &Detector{
		linguaAllLanguages:    lingua.NewLanguageDetectorBuilder().FromAllSpokenLanguages().WithPreloadedLanguageModels().Build(),
		linguaCommonLanguages: lingua.NewLanguageDetectorBuilder().FromLanguages(commonLanguages...).WithPreloadedLanguageModels().Build(),

		SelectDetectors: map[Channel]*SelectDetector{},
	}
}

func (d *Detector) ToJSON() ([]byte, error) {
	return json.Marshal(d)
}

func (d *Detector) FromJSON(jsonBytes []byte) error {
	var detector Detector
	if err := json.Unmarshal(jsonBytes, &detector); err != nil {
		return err
	}

	for channel, selectDetector := range detector.SelectDetectors {
		if _, err := d.UpdateSelected(
			string(channel),
			selectDetector.Selected.L1.String(),
			selectDetector.Selected.L2.String(),
		); err != nil {
			return err
		}
	}

	return nil
}

// Detect returns a best attempt at determining the input language.
// If the language can't be reliably detected, false is returned.
func (d *Detector) Detect(text string, threshold ...float32) (string, bool) {
	thresh := DefaultDetectionThreshold
	if len(threshold) > 0 {
		thresh = float64(threshold[0])
	}

	language, exists := d.linguaAllLanguages.DetectLanguageOf(text)
	if !exists {
		return "", false
	}

	confidence := d.linguaAllLanguages.ComputeLanguageConfidence(text, language)
	if confidence < thresh {
		language, exists = d.linguaCommonLanguages.DetectLanguageOf(text)
		if !exists {
			return "", false
		}
	}

	return language.String(), true
}

func (d *Detector) ClearSelected(channel string) {
	delete(d.SelectDetectors, Channel(channel))
}

func (d *Detector) UpdateSelected(channel, l1, l2 string) (bool, error) {
	// Determine language choices
	l1Lang := stringToLang(l1)
	l2Lang := stringToLang(l2)

	if l1Lang == lingua.Unknown || l2Lang == lingua.Unknown {
		return false, errors.New("unknown language selected")
	}

	// Retrieve select detector
	selectDetector, _ := d.GetSelectedDetector(channel)
	if selectDetector == nil {
		selectDetector = &SelectDetector{}
	}

	// Update?
	if selectDetector.Selected == nil || *selectDetector.Selected != (Selected{L1: l1Lang, L2: l2Lang}) {
		log.Printf("Reconfiguring select detector for %s:%s", l1Lang, l2Lang)
		selectDetector.linguaSelectLanguages =
			lingua.NewLanguageDetectorBuilder().FromLanguages([]lingua.Language{l1Lang, l2Lang}...).WithPreloadedLanguageModels().Build()
		selectDetector.Selected = &Selected{L1: l1Lang, L2: l2Lang}

		d.SelectDetectors[Channel(channel)] = selectDetector

		return true, nil
	}

	return false, nil
}

func (d *Detector) GetSelectedDetector(channel string) (*SelectDetector, error) {
	selectDetector, ok := d.SelectDetectors[Channel(channel)]
	if !ok {
		return nil, errors.New("channel not initialized for select detection")
	}
	return selectDetector, nil
}

// =========== Select Detector ============== //

func (s *SelectDetector) Select(channel, text string) (lingua.Language, error) {

	confidences := s.linguaSelectLanguages.ComputeLanguageConfidenceValues(text)
	sort.Slice(confidences, func(i, j int) bool {
		return confidences[i].Value() < confidences[j].Value()
	})

	return confidences[len(confidences)-1].Language(), nil
}

// =========== Helpers ================ //

func stringToLang(str string) lingua.Language {
	for _, l := range lingua.AllLanguages() {
		if strings.TrimSpace(strings.ToLower(l.String())) == strings.TrimSpace(strings.ToLower(str)) {
			return l
		}
	}
	return lingua.Unknown
}
