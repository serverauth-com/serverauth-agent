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
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/user"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync all ssh keys with ServerAuth",
	Long:  `This command will sync the authorized_keys file of each system account you have configured with ServerAuth.`,
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

		// Get the base domain, which can optionally be overridden
		var baseDomain string
		viper.UnmarshalKey("basedomain", &baseDomain)

		if len(baseDomain) <= 0 {
			// No overridden base domain, fall back to the default
			baseDomain = "https://api.serverauth.com/"
		}

		// Build the base url
		baseURL := baseDomain + "keys/" + orgId + "/" + serverAPIKey + "/"

		// Loop over accounts and sync
		for _, account := range accounts {

			// Check the user exists on the server, and save into a var for later use
			u, userErr := user.Lookup(account.Username)
			if userErr != nil {
				color.Red("Unable to find user `%s`. Please check the username, and re-create the user on ServerAuth.", username)
				return
			}

			accountAPIURL := baseURL + account.ApiKey
			color.Green("Loading API Key for %s from %s", account.Username, accountAPIURL)

			// Create http client
			httpClient := http.Client{
				Timeout: time.Second * 10, // Maximum of 10 secs
			}

			// Create a request
			req, err := http.NewRequest(http.MethodGet, accountAPIURL, nil)
			if err != nil {
				log.Fatal(err)
			}

			// Set our custom useragent
			req.Header.Set("User-Agent", "ServerAuthAgent-v2.0.0;"+runtime.GOOS)

			// Run the request
			res, getErr := httpClient.Do(req)
			if getErr != nil {
				log.Fatal(getErr)
			}

			// Process the body
			body, readErr := ioutil.ReadAll(res.Body)
			if readErr != nil {
				log.Fatal(readErr)
			}

			keys := string(body)

			// Validate that they keys file was valid
			validStart := strings.Contains(keys, "START ServerAuth Managed Keys File")
			validEnd := strings.Contains(keys, "END ServerAuth Managed Keys File")

			if !validStart || !validEnd {
				color.Red("The response from the ServerAuth api was invalid. Please contact us for assistance.")
				os.Exit(1)
			}

			// Work out the uid and gid for chowning
			uid, _ := strconv.Atoi(u.Uid)
			gid, _ := strconv.Atoi(u.Gid)

			// Keys for this user are valid. Save file
			// Set up our path and file vars
			homeDir := u.HomeDir
			keysDir := homeDir + "/.ssh"
			keysFile := keysDir + "/authorized_keys"
			color.Green("Writing to " + keysFile)

			// If the .ssh directory doesnt exist, create it and set it to be owned by the user
			if _, keysDirErr := os.Stat(keysDir); os.IsNotExist(keysDirErr) {
				color.Yellow("It looks like " + keysDir + " does not yet exist. Lets create it now.")
				os.MkdirAll(keysDir, 0700)
				os.Chown(keysDir, uid, gid)
			}

			// Ready to write the file
			ioutil.WriteFile(keysFile, []byte(keys), 0600)

			// Ensure the file is owned by the correct user
			os.Chown(keysFile, uid, gid)

			color.Green("Done!")
		}
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
