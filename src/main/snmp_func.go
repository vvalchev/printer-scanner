package main
import (
    "errors"
    "fmt"
    "net"
    "sync"
    "io"
  	"github.com/k-sone/snmpgo"
)

// https://docs.bmc.com/docs/display/Configipedia/List+of+discoverable+printers

//https://github.com/librenms/librenms-mibs/blob/master/IANA-PRINTER-MIB
//https://github.com/librenms/librenms-mibs/blob/master/Printer-MIB
//https://github.com/librenms/librenms-mibs/blob/master/SAMSUNG-PRINTER-EXT-MIB

//172.26.7.5

//snmpget -v1 -cpublic 172.26.7.5 1.3.6.1.2.1.1.1.0
// snmpwalk -v1 -cpublic 172.26.7.5
// https://exchange.nagios.org/directory/Plugins/Hardware/Printers/check_snmp_printer/details
// https://github.com/coreyramirezgomez/Brother-Printers-Zabbix-Template/blob/master/check_snmp_printer.sh
// https://github.com/PetrKohut/SNMP-printer-library/blob/master/library/Kohut/SNMP/Printer.php
// http://docs.sharpsnmp.com/en/latest/tutorials/device-discovery.html
//https://www.webnms.com/telecom/help/developer_guide/discovery/discovery_process/disc_broadcast.html

// https://en.wikipedia.org/wiki/Service_Location_Protocol
// http://jslp.sourceforge.net/

// https://developer.apple.com/bonjour/printing-specification/bonjourprinting-1.2.pdf

// https://serverfault.com/questions/154650/find-printers-with-nmap

// https://developer.android.com/reference/android/net/nsd/NsdManager.html
// https://sharpsnmplib.codeplex.com/wikipage?title=SNMP%20Device%20Discovery&referringTitle=Documentation



// http://www.snmplink.org/cgi-bin/nd/m/*25[All]%20Draft/Printer-MIB-printmib-04.txt
///usr/libexec/cups/backend/snmp
//HOST-RESOURCES-MIB::hrDeviceType.1 = OID: HOST-RESOURCES-TYPES::hrDevicePrinter
//HOST-RESOURCES-MIB::hrDeviceDescr.1 = STRING: HP LaserJet 4000 Series



////// PRINTER MIB:  snmpwalk -v1 -c public 172.26.7.5 1.3.6.1.2.1.43

////https://sourceforge.net/projects/mpsbox/?source=directory
// https://github.com/k-sone/snmpgo/blob/master/examples/snmpgobulkwalk.go


// https://github.com/apple/cups/blob/master/backend/backend-private.h
// http://oid-info.com/get/1.3.6.1.2.1.43
// http://www.ietf.org/rfc/rfc1759.txt
var CUPS_OID = []string{
    "1.3.6.1.2.1.1.1.0", // sysDescr
    "1.3.6.1.2.1.1.2.0", // sysObjectID
    "1.3.6.1.2.1.1.5.0", // sysName
    "1.3.6.1.2.1.43.5.1.1.17", // CUPS_OID_prtGeneralSerialNumber
    "1.3.6.1.2.1.43.5.1.1.16", // CUPS_OID_prtGeneralPrinterName
    "1.3.6.1.2.1.43.10", //CUPS_OID_prtMarker
}
// http://www.oidview.com/mibs/2590/MC2350-MIB.html
var MINOLTA_OID = []string{
    // mltSysDuplexCount 1.3.6.1.4.1.2590.1.1.1.5.7.2.1.3
    // mltSysTotalCount 1.3.6.1.4.1.2590.1.1.1.5.7.2.1.1
    "1.3.6.1.4.1.2590.1.1.1.5.7.2", // Minolta mltSysSystemCounter
    "1.3.6.1.4.1.18334.1.1.1.5.7.2.2.1.5", // Minolta MIB, kmSysPrintFunctionCounterTable
}
// http://www.oidview.com/mibs/11/IJXXXX-MIB.html
var HP_IJXXXX_OID = []string{
    // total-mono-page-count 1.3.6.1.4.1.11.2.3.9.4.2.1.4.1.2.6
    // total-color-page-count 1.3.6.1.4.1.11.2.3.9.4.2.1.4.1.2.7
    // duplex-page-count 1.3.6.1.4.1.11.2.3.9.4.2.1.4.1.2.22
    "1.3.6.1.4.1.11.2.3.9.4.2.1.4.1.2",
}
// http://www.oidview.com/mibs/2435/BROTHER-MIB.html
var BROTHER_OID = []string{
    // Brother printerinfomation
   "1.3.6.1.4.1.2435.2.3.9.4.2.1.5.5", // Brother printerinfomation
}
// https://github.com/librenms/librenms-mibs/blob/master/SAMSUNG-PRINTER-EXT-MIB
var SAMSUNG_OID = []string{
    "1.3.6.1.4.1.236.11.5.11.55.2.3.17",
    "1.3.6.1.4.1.236.11.5.11.55.2.3.20",
}
var EXTRA_OIDS = [][]string{
    MINOLTA_OID,
    HP_IJXXXX_OID,
    BROTHER_OID,
    SAMSUNG_OID,
}
var OID2PROP = map[string]string{
    "1.3.6.1.2.1.1.1.0": "sysDescr",
    "1.3.6.1.2.1.1.2.0": "sysObjectID",
    "1.3.6.1.2.1.1.5.0": "sysName",

    "1.3.6.1.2.1.43.5.1.1.17.1": "prtGeneralSerialNumber",
    "1.3.6.1.2.1.43.5.1.1.16.1": "prtGeneralPrinterName",

    "1.3.6.1.2.1.43.10.2.1.1.1.1": "prtMarkerIndex",
    "1.3.6.1.2.1.43.10.2.1.2.1.1": "prtMarkerMarkTech",
    "1.3.6.1.2.1.43.10.2.1.3.1.1": "prtMarkerCounterUnit",
    "1.3.6.1.2.1.43.10.2.1.4.1.1": "prtMarkerLifeCount",
    "1.3.6.1.2.1.43.10.2.1.5.1.1": "prtMarkerPowerOnCount",
    "1.3.6.1.2.1.43.10.2.1.15.1.1": "prtMarkerStatus",

    // hp
    "1.3.6.1.4.1.11.2.3.9.4.2.1.4.1.2.5.0": "total-engine-page-count",
    "1.3.6.1.4.1.11.2.3.9.4.2.1.4.1.2.6.0": "total-mono-page-count",
    "1.3.6.1.4.1.11.2.3.9.4.2.1.4.1.2.7.0": "total-color-page-count",
    "1.3.6.1.4.1.11.2.3.9.4.2.1.4.1.2.22.0": "duplex-page-count",

    // brother
    "1.3.6.1.4.1.2435.2.3.9.4.2.1.5.5.1":  "brInfoSerialNumber",
    "1.3.6.1.4.1.2435.2.3.9.4.2.1.5.5.10": "brInfoCounter",
    "1.3.6.1.4.1.2435.2.3.9.4.2.1.5.5.17": "brInfoDeviceRomVersion",

    // Minolta
    "1.3.6.1.4.1.18334.1.1.1.5.7.2.2.1.5.1.1": "copy-counter-black",
    "1.3.6.1.4.1.18334.1.1.1.5.7.2.2.1.5.2.1": "copy-counter-color",
    "1.3.6.1.4.1.18334.1.1.1.5.7.2.2.1.5.1.2": "print-counter-black",
    "1.3.6.1.4.1.18334.1.1.1.5.7.2.2.1.5.2.2": "print-counter-color",
}


func snmpScanOIDS(ip net.IP, oidsToScan []string) (varBinds snmpgo.VarBinds, err error)  {
    ip = ip.To4()
    if ip == nil {
        return nil, errors.New("Scan IP failed - IP is not specified!") // not IPv4 address
    }

    snmp, err := snmpgo.NewSNMP(snmpgo.SNMPArguments{
        Version:   snmpgo.V2c,
        Address:   fmt.Sprintf("%v:161", ip),
        Retries:   1,
        Community: "public",
    })
    if err != nil {
        // log.Printf("Failed to allocate SNMP Request: %v\n", err)
        return nil, err
    }

    oids, err := snmpgo.NewOids(oidsToScan)/*

    }*/
    if err != nil {
        // log.Printf("Failed to parse Oids: %v\n", err)
        return nil, err
    }

    if err = snmp.Open(); err != nil {
        // log.Printf("Failed to open connection to %v: %v\n", ip, err)
        return
    }
    defer snmp.Close()

    // pdu, err := snmp.GetRequest(oids)
    var nonRepeaters = 0
    var maxRepetitions = 10
    pdu, err := snmp.GetBulkWalk(oids, nonRepeaters, maxRepetitions)
    if err != nil {
        return nil, err
    }

    return pdu.VarBinds(), nil
}

func SnmpScan(ip net.IP) (varBinds snmpgo.VarBinds, err error) {
    // check if it is a CUPS printer
    vBinds,err := snmpScanOIDS(ip, CUPS_OID)
    if err != nil {
        return nil, err
    }

    // asynchronously scan for specific manufacturer
    var wg sync.WaitGroup
    for _,oids := range EXTRA_OIDS {
        wg.Add(1)
        go func (xip net.IP, xoidx []string) {
            defer wg.Done()
            vBinds2,_ := snmpScanOIDS(xip, xoidx)
            if vBinds2 != nil {
                vBinds = append(vBinds, vBinds2...)
            }
        }(ip, oids)
    }

    wg.Wait()
    return vBinds, nil
}

func SnmpPrint(w io.Writer, ip net.IP, vars snmpgo.VarBinds) {
    fmt.Fprintf(w, "[%v]\n", ip)
    if vars != nil {
        for _, val := range vars {
            oidS := val.Oid.String()
            prop,ok := OID2PROP[oidS]
            if !ok {
                prop = oidS
            }
            fmt.Fprintf(w, "%s = %s\n", prop, val.Variable)
        }
    }
    fmt.Fprintf(w, "\n\n")
}


type JsonVar struct {
    Key string
    Value string
}
type JsonVars struct {
    Ip string
    Data map[string]string

}
func Snmp2Json(ip net.IP, vars snmpgo.VarBinds) (ret JsonVars) {
    ret = JsonVars{
        Ip: ip.String(),
        Data: make(map[string]string),
    }
    if vars != nil {
        for _, val := range vars {
            oidS := val.Oid.String()
            prop,ok := OID2PROP[oidS]
            if !ok {
                prop = oidS
            }
            ret.Data[prop] = val.Variable.String()
        }
    }
    return ret
}
