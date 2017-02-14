package main

import "flag"

func main() {
	samplesPtr := flag.Int("samples", 100, "How many DNS Requests you want to make")
	domainPtr := flag.String("domain", "www.google.com", "The domain you wish to resolve.")
	serverPtr := flag.String("server", "8.8.8.8", "The DNS Server you wish to query against.")

	flag.Parse()

	analyzeDns(*serverPtr, *domainPtr, *samplesPtr)
}
