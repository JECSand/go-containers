package containers

import (
	"strings"
	"time"
)

// getTimeStamp
func getTimeStamp() string {
	currentTime := time.Now().UTC()
	t := strings.Split(currentTime.String(), " +")[0]
	t = strings.Replace(t, " ", "T", 1)
	t = strings.Replace(t, " ", "", -1)
	t = t + "-UTC"
	return t
}
