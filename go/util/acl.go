package util

import (
	"errors"
	"net"
	"regexp"
)

type aclEntry struct {
	ipnet *net.IPNet
	allow bool
}

type ACL struct {
	acl []aclEntry
}

func parseACLEntry(v string) (aclEntry, error) {
	var allow = true
	if v[0] == '+' || v[0] == '-' {
		allow = v[0] == '+'
		v = v[1:]
	}
	_, ipnet, err := net.ParseCIDR(v)
	if err != nil {
		return aclEntry{}, err
	}
	return aclEntry{ipnet, allow}, nil
}

func ParseACL(s string) ([]aclEntry, error) {
	var l = speRegexp.Split(s, -1)
	var acl []aclEntry
	for _, v := range l {
		if len(v) > 6 {
			e, err := parseACLEntry(v)
			if err != nil {
				return nil, err
			}
			acl = append(acl, e)
		} else {
			return nil, errors.New("Invalid ACL: " + v)
		}
	}
	return acl, nil
}

var speRegexp = regexp.MustCompile(`\s+`)

func (a *ACL) IsAllowed(ip net.IP) bool {
	for _, e := range a.acl {
		if e.ipnet.Contains(ip) {
			return e.allow
		}
	}
	return false
}
