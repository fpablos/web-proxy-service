package couchbase

import (
	"time"
)

type PathTryed struct {
	Path					string		`json:"path"`
	Successful				int64		`json:"successful"`
	Rejected 				int64		`json:"rejected"`
	Date					Time 		`json:"date"`
}

type Host struct {
	Ip						string		`json:"ip"`
	Successful				int64		`json:"successful"`
	Rejected 				int64		`json:"rejected"`
	Date					Time 		`json:"date"`
}

type HostStatistic struct{
	Ip 						string		`json:"ip"`
	Successful				int64		`json:"successful"`
	Rejected 				int64		`json:"rejected"`
	Hosts					[]Host		`json:"hosts"`
	Paths					[]PathTryed	`json:"paths"`
}

// The time is expected to be a quoted string in RFC 3339 format.
type Time struct{
	time.Time
}

func (t Time) MarshalJSON() ([]byte, error) {
	return []byte(t.Format(`"`+time.RFC3339+`"`)), nil
}

func (t *Time) UnmarshalJSON(data []byte) (err error) {
	if string(data) == "null" || string(data) == `""` {
		return nil
	}

	tt, err := time.Parse(`"`+time.RFC3339+`"`, string(data))
	*t = Time{tt}
	return
}

func (hs HostStatistic) ConnectionsCountByIpSuccessful(destIp string) (int64, bool) {
	for _, host := range hs.Hosts {
		if host.Ip == destIp {
			return host.Successful, true
		}
	}

	return 0, false
}

func (hs HostStatistic) ConnectionsCountByPathSuccessful (destPath string) (int64, bool) {
	for _, path := range hs.Paths {
		if path.Path == destPath {
			return path.Successful, true
		}
	}

	return 0, false
}