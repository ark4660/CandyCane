package transcript

import (

	"strings"
	"encoding/json"
	"bytes"
	"io"
	"log"
	"strconv"
	"bufio"

)
type Segment struct {
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	Text  string  `json:"text"`
}

type Transcript struct {
	Segments []Segment `json:"segments"`
}


func vttTimestampToSeconds(ts string) float64 {
	ts = strings.TrimSpace(ts)
	parts := strings.Split(ts, ":")

	var h, m float64
	var s float64

	if len(parts) == 3 {
		h, _ = strconv.ParseFloat(parts[0], 64)
		m, _ = strconv.ParseFloat(parts[1], 64)
		s, _ = strconv.ParseFloat(strings.Replace(parts[2], ",", ".", 1), 64)
	} else {
		m, _ = strconv.ParseFloat(parts[0], 64)
		s, _ = strconv.ParseFloat(strings.Replace(parts[1], ",", ".", 1), 64)
	}

	return h*3600 + m*60 + s
}

func VttToJSON(file io.Reader) (io.ReadCloser, error) {
	var segments []Segment
	var currentText []string
	var currentStart, currentEnd float64
	inBlock := false

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip header
		if line == "WEBVTT" || line == "" {
			if inBlock && len(currentText) > 0 {
				segments = append(segments, Segment{
					Start: currentStart,
					End:   currentEnd,
					Text:  strings.Join(currentText, " "),
				})
				currentText = nil
				inBlock = false
			}
			continue
		}

		// Timestamp line
		if strings.Contains(line, "-->") {
			parts := strings.Split(line, "-->")
			currentStart = vttTimestampToSeconds(parts[0])
			currentEnd = vttTimestampToSeconds(parts[1])
			inBlock = true
			continue
		}

		// Skip cue identifiers (pure numbers)
		if _, err := strconv.Atoi(line); err == nil {
			continue
		}

		// Text content
		if inBlock {
			currentText = append(currentText, line)
		}
	}

	// Catch last block
	if inBlock && len(currentText) > 0 {
		segments = append(segments, Segment{
			Start: currentStart,
			End:   currentEnd,
			Text:  strings.Join(currentText, " "),
		})
	}

	transcript := Transcript{Segments: segments}
	out, err := json.MarshalIndent(transcript, "", "  ")
	if err != nil {
		log.Println(err)
	}
	reader := bytes.NewReader(out)
	readCloser := io.NopCloser(reader)
	return readCloser, err
}
