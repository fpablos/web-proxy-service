package couchbase

type Blacklist struct {
	Ip 				[]string		`json:"ip"`
}

func (b Blacklist) IpIsInBlackList(ip string) bool{
	if contains(b.Ip, ip) {
		return true
	}
	return false
}

// Contains tells whether a contains x.
func contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}


