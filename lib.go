package main

import (
	"fmt"
	"io"
	"log"
	"math"
	"sort"
	"time"

	"github.com/miekg/dns"
)

type DurationSlice []time.Duration

// NOTE: This implements the sortable interface
func (p DurationSlice) Len() int           { return len(p) }
func (p DurationSlice) Less(i, j int) bool { return int64(p[i]) < int64(p[j]) }
func (p DurationSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// NOTE: Wasteful Convenience Functions
func (p DurationSlice) Min() time.Duration {
	sort.Sort(p)
	return p[0]
}
func (p DurationSlice) Max() time.Duration {
	sort.Sort(p)
	return p[p.Len()-1]
}
func (p DurationSlice) Avg() time.Duration {
	var avg int64
	for i := 0; i < p.Len(); i++ {
		avg += int64(p[i])
	}
	return time.Duration(avg / int64(p.Len()))
}
func (p DurationSlice) Std() time.Duration {
	sqdifs := make(DurationSlice, p.Len(), p.Len())
	avg := p.Avg()
	var avgsqdif int64
	for i := 0; i < p.Len(); i++ {
		sqdif := p[i] - avg
		sqdifs[i] = sqdif * sqdif
		avgsqdif += int64(sqdifs[i])
	}
	avgsqdif /= int64(sqdifs.Len())
	return time.Duration(math.Sqrt(float64(avgsqdif)))
}

// TODO(zenware): make this output to a writer
func analyzeDns(w io.Writer, server, hostname string, samples, waitMillis int) {
	m := new(dns.Msg)
	m.Id = dns.Id()
	m.RecursionDesired = true
	m.Question = make([]dns.Question, 1)
	m.Question[0] = dns.Question{Name: dns.Fqdn(hostname), Qtype: dns.TypeA, Qclass: dns.ClassINET}
	wait := time.Duration(waitMillis) * time.Millisecond

	c := new(dns.Client)

	fmt.Printf("QUERY %v (@%v):  %v data bytes\n", hostname, server, m.Len())

	rtts := make(DurationSlice, samples, samples)
	for i := 0; i < samples; i++ {
		in, rtt, err := c.Exchange(m, server+":53")
		if err != nil {
			log.Println(err)
			continue
		}
		rtts[i] = rtt
		fmt.Fprintf(w, "%v bytes from %v: ttl=%v time=%v\n", in.Len(), server, time.Second*6, rtt)
		time.Sleep(wait)
	}

	// NOTE: Potentially Eating Performance for Pretties
	var min, max, avg, stddev time.Duration
	min = rtts.Min()
	max = rtts.Max()
	avg = rtts.Avg()
	stddev = rtts.Std()

	fmt.Fprintf(w, "round-trip min/avg/max/stddev = %v/%v/%v/%v\n", min, avg, max, stddev)
}
