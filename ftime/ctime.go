package ftime

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"time"
)

// type Ftime interface {
// 	ToStr() string
// 	WithTimeStr() Ftime
// }

type CTime struct {
	time.Time
}

func (j *CTime) Parse(str string) error {
	lstlayout := []string{"2006-01-02T15:04:05"}

	for _, layout := range lstlayout {
		if len(str) == len(layout) {

			t, err := time.Parse(layout, str)
			if err == nil {
				*j = CTime{Time: t}
			}
			break
		}
	}
	return nil
}
func (j *CTime) UnmarshalJSON(b []byte) error {
	*j = CTime{}.WithTimeStr(string(b))
	return nil
}

func (t CTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.ToStr())
}

func (t CTime) ToStr() string {
	unix := t.UnixMilli()
	if t.IsZero() {
		unix = 0
	}
	// return fmt.Sprintf("/Date(%d)/", unix)
	return fmt.Sprintf("%d", unix)
}

func (t CTime) WithTimeStr(s string) CTime {
	re := regexp.MustCompile("[0-9]+")
	unix, _ := strconv.ParseInt(re.FindString(s), 10, 64)
	t.Time = time.UnixMilli(unix)
	return t
}
