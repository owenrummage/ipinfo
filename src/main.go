package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"

	"strings"
	"text/template"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

type AddressInformation struct {
	Ip       string
	City     string
	Region   string
	Country  string
	Log      string
	Org      string
	Postal   string
	Timezone string
}

func format(s string, v interface{}) string {
	t, b := new(template.Template), new(strings.Builder)
	template.Must(t.Parse(s)).Execute(b, v)
	return b.String()
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func IsIPv4(address string) bool {
	return strings.Count(address, ":") < 2
}

func IsIPv6(address string) bool {
	return strings.Count(address, ":") >= 2
}

func main() {
	app := &cli.App{
		Name:  "ipinfo",
		Usage: "Get IP address information.",
		Commands: []*cli.Command{
			{
				Name:    "address",
				Aliases: []string{"a"},
				Usage:   "get local IP address information",
				Action: func(cCtx *cli.Context) error {
					ifaces, err := net.Interfaces()
					if err != nil {
						fmt.Print(fmt.Errorf("localAddresses: %+v\n", err.Error()))
						return nil
					}

					resp, httpErr := http.Get("https://api.ipify.org")
					if httpErr == nil {
						defer resp.Body.Close()
						body, ioErr := io.ReadAll(resp.Body)
						if ioErr == nil {
							fmt.Println(color.MagentaString("PUBLIC ADDRESS:") + " " + color.WhiteString(string(body)))
						}
					}

					fmt.Println(color.MagentaString("LOCAL INTERFACES:"))
					for _, i := range ifaces {
						if i.Name[:2] == "lo" {
							continue
						}
						if len(i.Name) >= 6 && i.Name[:6] == "docker" {
							continue
						}

						var v4Addr, v6Addr string

						addrs, err := i.Addrs()
						if err != nil {
							fmt.Print(fmt.Errorf("localAddresses: %+v\n", err.Error()))
							continue
						}
						for _, a := range addrs {
							if a.String() == "" || a.String()[:4] == "fe80" {
								continue
							}

							if IsIPv4(a.String()) {
								v4Addr = a.String()
							}
							if IsIPv6(a.String()) {
								v6Addr = a.String()
							}
						}

						if v4Addr == "" && v6Addr == "" {
							continue
						}
						fmt.Printf(color.GreenString("  "+i.Name+": ")+" %s (%s)\n", v4Addr, v6Addr)
					}
					return nil
				},
			},
			{
				Name:    "lookup",
				Aliases: []string{"l"},
				Usage:   "lookup an IP address",
				Action: func(cCtx *cli.Context) error {
					if cCtx.Args().First() == "" {
						fmt.Println(color.RedString("IP address argument is required."))
						return nil
					}

					ip := net.ParseIP(cCtx.Args().First())
					if ip == nil {
						fmt.Println(color.RedString("That is not a valid IP Address."))
						return nil
					}

					resp, httpErr := http.Get("https://ipinfo.io/" + cCtx.Args().First())

					if httpErr != nil {
						fmt.Println(color.RedString(httpErr.Error()))
						return nil
					}

					body, ioErr := io.ReadAll(resp.Body)

					if ioErr != nil {
						fmt.Println(color.RedString(ioErr.Error()))
						return nil
					}

					var infoObject AddressInformation

					jsonErr := json.Unmarshal(body, &infoObject)

					if jsonErr != nil {
						fmt.Println(color.RedString(jsonErr.Error()))
						return nil
					}

					fmt.Printf(format(color.MagentaString("IPINFO - Address Information")+"\n  Address: {{.Ip}}\n  Location: {{.City}}, {{.Region}} {{.Country}} ({{.Postal}})\n  Organization: {{.Org}}", infoObject))
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
