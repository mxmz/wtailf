package util

import (
	"net"
	"testing"

	. "github.com/onsi/gomega"
)

func TestparseACLEntry(t *testing.T) {
	g := NewWithT(t)
	{
		sut, err := parseACLEntry("192.168.0.33/27") // non canonical, should be .32
		g.Expect(err).To(BeNil())
		g.Expect(sut.allow).To(BeTrue())
		g.Expect(sut.ipnet.IP).To(BeEquivalentTo([]byte{192, 168, 0, 32}))
		g.Expect(sut.ipnet.Mask).To(BeEquivalentTo([]byte{255, 255, 255, 224}))
	}
	{
		_, err := parseACLEntry("-192.168.0.33/27")
		g.Expect(err).ToNot(BeNil())
	}
}

func TestParseACL(t *testing.T) {
	g := NewWithT(t)
	var err error
	sut, err := ParseACL("192.168.0.33/27 +194.168.0.33/27 -198.168.0.33/27")
	g.Expect(err).To(BeNil())
	g.Expect(sut[0].allow).To(BeTrue())
	g.Expect(sut[0].ipnet.IP).To(BeEquivalentTo([]byte{192, 168, 0, 32}))
	g.Expect(sut[0].ipnet.Mask).To(BeEquivalentTo([]byte{255, 255, 255, 224}))
	g.Expect(sut[1].allow).To(BeTrue())
	g.Expect(sut[1].ipnet.IP).To(BeEquivalentTo([]byte{194, 168, 0, 32}))
	g.Expect(sut[1].ipnet.Mask).To(BeEquivalentTo([]byte{255, 255, 255, 224}))
	g.Expect(sut[2].allow).To(BeFalse())
	g.Expect(sut[2].ipnet.IP).To(BeEquivalentTo([]byte{198, 168, 0, 32}))
	g.Expect(sut[2].ipnet.Mask).To(BeEquivalentTo([]byte{255, 255, 255, 224}))
}

func TestACL_IsAllowed(t *testing.T) {
	g := NewWithT(t)
	var err error
	acl, err := ParseACL("192.168.0.33/27 +194.168.0.33/27 -198.168.0.33/27")
	g.Expect(err).To(BeNil())
	sut := ACL{acl}
	g.Expect(sut.IsAllowed(net.ParseIP("194.168.0.39"))).To(BeTrue())
	g.Expect(sut.IsAllowed(net.ParseIP("198.168.0.39"))).To(BeFalse())
}
