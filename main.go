package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
)

var doScan = flag.Bool("scan", false, "Scan the complete local network to find Printer Devices.")
var outFile = flag.String("o", "", "The output file, where the scan results to be written.")
var url = flag.String("post", "", "Optional URL. When specified the printer data is serialized to JSON and posted to that URL.")
var clientId = flag.String("clientId", "", "Option clientId, that is supplied when posting printer data to URL.")

type PostData struct {
	ClientId string
	Printers []JsonVars
}

func postPrinterData(d PostData) {
	log.Printf("Publishing printer data to URL: %s\n", *url)

	// serialize to JSON
	jsonString, _ := json.Marshal(d)

	// POST
	resp, err := http.Post(*url, "application/json", bytes.NewBuffer(jsonString))
	if err != nil {
		log.Printf("Failed to publish printer data: %v\n", err)
		return
	}
	defer resp.Body.Close()
	log.Printf("Publishing printer data to URL '%s' completed with code: %s\n", *url, resp.Status)
}

func main() {
	flag.Parse()

	var ipList []net.IP
	out := os.Stdout

	if outFile != nil && *outFile != "" {
		var err error
		out, err = os.Create(*outFile)
		if err != nil {
			log.Printf("Cannot open file %v for writing, using console instead!\n", outFile)
			defer out.Close()
			out = os.Stdout
			*outFile = "console"
		}
	}

	if *doScan {
		// list of IPs to scan, based on current network interfaces
		ipList, _ = NetGetNetworkIPs()
	}

	// list of printers added by command line
	for _, arg := range flag.Args() {
		ips, _ := net.LookupIP(arg)
		for _, ip := range ips {
			ipList = append(ipList, ip)
		}
	}

	if ipList != nil {
		json := PostData{
			ClientId: *clientId,
			Printers: []JsonVars{},
		}

		log.Printf("Scanning started ...\n")
		// asynchronously scan every IP address
		var wg sync.WaitGroup
		for _, ip := range ipList {
			wg.Add(1)
			go func(xip net.IP) {
				defer wg.Done()
				vars, _ := SnmpScan(xip)
				if vars != nil {
					log.Printf("%16v -> OK\n", xip)
					json.Printers = append(json.Printers, Snmp2Json(xip, vars))
					SnmpPrint(out, xip, vars)
				} else {
					log.Printf("%16v -> FAIL\n", xip)
				}
			}(ip)
		}
		wg.Wait()
		if *url != "" {
			// ipVars.ClientId = *clientId
			postPrinterData(json)
		}
		log.Printf("Scanning complete! Results are printed to %v\n", *outFile)
	} else {
		log.Printf("No printers to scan, please provide a list of printers or -scan parameter!\n")
	}

}
