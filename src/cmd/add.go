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
	"log"
	"os"
	"os/user"
	"strconv"

	"github.com/spf13/viper"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var username string
var apikey string

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a system account to ServerAuth",
	Long:  `Add a new system account (e.g root) to ServerAuth to have it's SSH Keys automatically managed.`,
	Run: func(cmd *cobra.Command, args []string) {

		// Check the user exists on the server
		u, err := user.Lookup(username)
		if err != nil {
			color.Red("Unable to find user `%s`. Please check the username, and re-create the user on ServerAuth.", username)
			return
		}

		color.Green("Found system user: %s\nSetting up ServerAuth for the account.", u.Username)

		// Read in the existing accounts and get ready for adding another user
		viper.ReadInConfig()

		var accounts []Account
		configErr := viper.UnmarshalKey("accounts", &accounts)

		if configErr != nil {
			panic("There was a problem setting up the user account. Please try again or contact ServerAuth for assistance")
		}

		// Append the new account and update the config
		accounts = append(accounts, Account{username, apikey})
		viper.Set("accounts", accounts)
		viper.WriteConfig()

		// Work out the uid and gid for chowning
		uid, _ := strconv.Atoi(u.Uid)
		gid, _ := strconv.Atoi(u.Gid)

		// Set up our path and file vars
		homeDir := u.HomeDir
		keysDir := homeDir + "/.ssh"
		keysFile := keysDir + "/authorized_keys"
		backupKeysFile := keysDir + "/.ssh/authorized_keys.bak"

		// If the .ssh directory doesnt exist, create it and set it to be owned by the user
		if _, keysDirErr := os.Stat(keysDir); os.IsNotExist(keysDirErr) {
			color.Yellow("It looks like " + keysDir + " does not yet exist. Lets create it now.")
			os.MkdirAll(keysDir, 0700)
			os.Chown(keysDir, uid, gid)
		}

		// If we find an authorized_keys file, move it to a backup location so we dont overwrite it.
		if _, keysFileErr := os.Stat(keysFile); !os.IsNotExist(keysFileErr) {
			// File exists, move it to its backup location
			backupFileErr := os.Rename(keysFile, backupKeysFile)
			if backupFileErr != nil {
				log.Fatal(backupFileErr)
			}
			color.Yellow("An existing authorized_keys file was found. This has been moved to %s", backupKeysFile)

		}

		// Create our authorized_keys file
		file, fileError := os.Create(keysFile)
		if fileError != nil {
			panic("Unable to create authorized_keys file. Please check that the user you are running the agent as has the correct privieges.")
		}

		// Write the template to the file
		file.Write(keysFileTemplate)

		// Ensure the file is owned by the correct user
		os.Chown(keysFile, uid, gid)

		color.Green("The user was successfully configured and is now managed by ServerAuth.")
	},
}

func init() {
	rootCmd.AddCommand(addCmd)

	// User flag
	addCmd.Flags().StringVarP(&username, "username", "u", "", "The username of the system account to add to ServerAuth")
	addCmd.MarkFlagRequired("user")

	// API Key Flag
	addCmd.Flags().StringVarP(&apikey, "apikey", "k", "", "The unique API Key for the system account, provided when adding the account via your ServerAuth control panel.")
	addCmd.MarkFlagRequired("api-key")
}
