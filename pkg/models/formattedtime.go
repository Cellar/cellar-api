package models

import (
	"encoding/json"
	"time"
)

type FormattedTime time.Time

var layout = "2006-01-02 15:04:05 MST"

func (t FormattedTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Format())
}

func (t FormattedTime) Format() string {
	return time.Time(t).Format(layout)
}

func (t *FormattedTime) UnmarshalJSON(data []byte) error {
	var formattedStr string
	err := json.Unmarshal(data, &formattedStr)
	if err != nil {
		return err
	}

	tm, err := time.Parse(layout, formattedStr)
	if err != nil {
		return err
	}

	*t = FormattedTime(tm)

	return nil
}

func (t FormattedTime) Time() time.Time {
	return time.Time(t)
}
