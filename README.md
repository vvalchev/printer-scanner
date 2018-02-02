# Introduction
This is a simple project, that scans network for printers (via SNMP) and retrieves some basic information including page counters.

# Building the project
To build the project you need to have docker installed (or go lang).

In case you are building with docker, you have to run ```./build.sh```

If you already have go lang installed you build with ```./build.sh build```

64bit binaries for the following platforms will be created:
* Linux
* MacOS
* Window

# Usage
```
Usage of ./printer-scanner:
  -clientId string
    	Option clientId, that is supplied when posting printer data to URL.
  -o string
    	The output file, where the scan results to be written. (default "printers.ini")
  -post string
    	Optional URL. When specified the printer data is serialized to JSON and posted to that URL.
  -scan
    	Scan the complete local network to find Printer Devices.```

If you omit the '-scan' you have to provide a list of IP/Network Addresses to scan.

In case you specify a POST URL, the format of the data that is send that URL is JSON:
```
{
    "ClientId": "my-very-secret-client-id",
    "Printers": [
        {
            "Ip": "172.26.7.5",
            "Data": {
                "1.3.6.1.2.1.43.10.2.1.10.1.1": "600",
                "duplex-page-count": "3343",
                "prtGeneralPrinterName": "HP LaserJet Pro MFP M225dw",
                "prtGeneralSerialNumber": "CNB8G96C25",
                "prtMarkerCounterUnit": "7",
                "prtMarkerLifeCount": "10339",
                "prtMarkerMarkTech": "4",
                "prtMarkerPowerOnCount": "10",
                "prtMarkerStatus": "0",
                "sysName": "NPI5D217D",
                "sysObjectID": "1.3.6.1.4.1.11.2.3.9.1",
                "total-engine-page-count": "10339",
                "total-mono-page-count": "10339"
            }
        }
    ]
}
```

The number of the keys will be different depending on the printer model or manufacturer.
Some OIDs would be included with their OID name. Others, which are understood by this app will
be translated to some meaningful names.
