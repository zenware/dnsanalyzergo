package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"sort"
	"time"

	"github.com/miekg/dns"
)

type DurationSlice []time.Duration

func (p DurationSlice) Len() int           { return len(p) }
func (p DurationSlice) Less(i, j int) bool { return int64(p[i]) < int64(p[j]) }
func (p DurationSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func main() {
	samplesPtr := flag.Int("samples", 100, "How many DNS Requests you want to make")
	domainPtr := flag.String("domain", "www.google.com", "The domain you wish to resolve.")
	serverPtr := flag.String("server", "8.8.8.8", "The DNS Server you wish to query against.")

	flag.Parse()

	m := new(dns.Msg)
	m.Id = dns.Id()
	m.RecursionDesired = true
	m.Question = make([]dns.Question, 1)
	m.Question[0] = dns.Question{dns.Fqdn(*domainPtr), dns.TypeA, dns.ClassINET}

	c := new(dns.Client)

	fmt.Printf("QUERY %v (@%v):  %v data bytes\n", *domainPtr, *serverPtr, m.Len())

	rtts := make(DurationSlice, *samplesPtr, *samplesPtr)
	for i := 0; i < *samplesPtr; i++ {
		in, rtt, err := c.Exchange(m, *serverPtr+":53")
		if err != nil {
			log.Println(err)
		}
		rtts[i] = rtt
		fmt.Printf("%v bytes from %v: ttl=%v time=%v\n", in.Len(), *serverPtr, time.Second*6, rtt)
	}
	sort.Sort(rtts)

	var min, max, avg, avgsqdif, stdev time.Duration
	min = time.Duration(rtts[0])
	max = time.Duration(rtts[len(rtts)-1])
	avg = (min + max) / 2

	sqdifs := make(DurationSlice, *samplesPtr, *samplesPtr)
	for j := 0; j < *samplesPtr; j++ {
		sqdif := rtts[j] - avg
		sqdifs[j] = sqdif * sqdif
	}
	sort.Sort(sqdifs)

	avgsqdif = (sqdifs[0] + sqdifs[len(sqdifs)-1]) / 2
	stdev = time.Duration(math.Sqrt(float64(avgsqdif)))

	fmt.Printf("round-trip min/avg/max/stddev = %v/%v/%v/%v\n", min, avg, max, stdev)
}
