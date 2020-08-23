package util

import (
	"errors"
	"net"
	"regexp"
)

type ACLEntry struct {
	ipnet *net.IPNet
	allow bool
}

type ACL struct {
	acl []ACLEntry
}

func NewACL(acl ...ACLEntry) ACL {
	return ACL{acl}
}
func NewACLEntry(n *net.IPNet, allow bool) ACLEntry {
	return ACLEntry{n, allow}
}

func LocalhostAllow() ACLEntry {
	_, ipnet, _ := net.ParseCIDR("127.0.0.0/8")
	return ACLEntry{ipnet, true}
}

func parseACLEntry(v string) (ACLEntry, error) {
	var allow = true
	if v[0] == '+' || v[0] == '-' {
		allow = v[0] == '+'
		v = v[1:]
	}
	_, ipnet, err := net.ParseCIDR(v)
	if err != nil {
		return ACLEntry{}, err
	}
	return ACLEntry{ipnet, allow}, nil
}

func ParseACL(s string) ([]ACLEntry, error) {
	var l = speRegexp.Split(s, -1)
	var acl []ACLEntry
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
