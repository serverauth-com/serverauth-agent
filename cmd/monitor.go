/*
Copyright © 2019 ServerAuth.com <info@serverauth.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"

	"github.com/shirou/gopsutil/mem"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var actionCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Collect server metrics",
	Long:  `Collects the latest server monitoring metrics and sends them to your ServerAuth account.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Read in the existing accounts and get ready for adding another user
		viper.ReadInConfig()

		var accounts []Account
		configErr := viper.UnmarshalKey("accounts", &accounts)

		if configErr != nil {
			color.Red("There was a problem setting up the user account.\nPlease try again or contact ServerAuth for assistance.")
			os.Exit(1)
		}

		// Get the organisation id
		var orgId string
		viper.UnmarshalKey("orgid", &orgId)

		if len(orgId) <= 0 {
			color.Red("The organisation id is missing from your ServerAuth configuration.\nPlease check you've correctly configured ServerAuth on this server and try again.")
			os.Exit(1)
		}

		// Get the server api key
		var serverAPIKey string
		viper.UnmarshalKey("apikey", &serverAPIKey)

		if len(serverAPIKey) <= 0 {
			color.Red("The server API key is missing.\nPlease check you've correctly configured ServerAuth on this server and try again.")
			os.Exit(1)
		}

		// Get the team api key
		var teamAPIKey string
		viper.UnmarshalKey("teamkey", &teamAPIKey)

		if len(teamAPIKey) <= 0 {
			color.Red("The team API key is missing.\nPlease check you've correctly configured ServerAuth on this server and try again.")
			os.Exit(1)
		}

		// Get the base domain, which can optionally be overridden
		var baseDomain string
		viper.UnmarshalKey("basedomain", &baseDomain)

		if len(baseDomain) <= 0 {
			// No overridden base domain, fall back to the default
			baseDomain = "https://api.serverauth.com/"
		}

		// Build the base url
		baseURL := baseDomain + "monitoring"

		// Get current system stats
		memory, _ := mem.VirtualMemory()
		loadAvg, _ := load.Avg()
		cpuHw, _ := cpu.Info()
		//percent, _ := cpu.Percent(time.Second, true)
		uptime, _ := host.Uptime()
		misc, _ := load.Misc()
		platform, family, version, _ := host.PlatformInformation()
		diskStat, err := disk.Usage("/")
		currentTime := time.Now()
		timeZone, timeOffset := currentTime.Zone()

		// Create http client
		httpClient := http.Client{
			Timeout: time.Second * 10, // Maximum of 10 secs
		}

		form := url.Values{}

		// Mem
		form.Add("mem[total]", fmt.Sprint(memory.Total))
		form.Add("mem[free]", fmt.Sprint(memory.Free))
		form.Add("mem[used]", fmt.Sprint(memory.Used))
		form.Add("mem[used_percent]", fmt.Sprint(memory.UsedPercent))

		// Loadavg
		form.Add("load[1]", fmt.Sprint(loadAvg.Load1))
		form.Add("load[5]", fmt.Sprint(loadAvg.Load5))
		form.Add("load[15]", fmt.Sprint(loadAvg.Load15))

		// Processes
		form.Add("procs[running]", fmt.Sprint(misc.ProcsRunning))
		form.Add("procs[blocked]", fmt.Sprint(misc.ProcsBlocked))
		form.Add("procs[total]", fmt.Sprint(misc.ProcsTotal))

		// Generic CPU HW
		form.Add("cpu", fmt.Sprint(cpuHw))

		// Uptime
		form.Add("uptime_seconds", fmt.Sprint(uptime))

		// Platform info
		form.Add("platform[name]", fmt.Sprint(platform))
		form.Add("platform[family]", fmt.Sprint(family))
		form.Add("platform[version]", fmt.Sprint(version))

		// Disk info
		form.Add("disk[total]", strconv.FormatUint(diskStat.Total, 10))
		form.Add("disk[used]", strconv.FormatUint(diskStat.Used, 10))
		form.Add("disk[free]", strconv.FormatUint(diskStat.Free, 10))
		form.Add("disk[percent_free]", strconv.FormatFloat(diskStat.UsedPercent, 'f', 2, 64))
		form.Add("disk[inodes_total]", strconv.FormatUint(diskStat.InodesTotal, 10))
		form.Add("disk[inodes_used]", strconv.FormatUint(diskStat.InodesUsed, 10))
		form.Add("disk[inodes_free]", strconv.FormatUint(diskStat.InodesFree, 10))

		// Time info
		form.Add("time[zone]", fmt.Sprint(timeZone))
		form.Add("time[offset]", fmt.Sprint(timeOffset))
		form.Add("time[now]", fmt.Sprint(currentTime.Unix()))

		// Create a request
		req, err := http.NewRequest(http.MethodPost, baseURL, strings.NewReader(form.Encode()))
		if err != nil {
			log.Fatal(err)
		}

		req.Header.Set("User-Agent", "ServerAuthAgent-v2.0.0;"+runtime.GOOS)
		req.Header.Set("TeamApiKey", teamAPIKey)
		req.Header.Set("ServerApiKey", serverAPIKey)
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		httpClient.Do(req)
	},
}

func init() {
	rootCmd.AddCommand(actionCmd)
}
