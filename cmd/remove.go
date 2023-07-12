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
	"os/user"
    "fmt"
    "time"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// removeCmd represents the remove command
var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a system account from ServerAuth",
	Long:  `Remove a system account from ServerAuth's automatic SSH Key management.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check the user exists on the server
		u, err := user.Lookup(username)
		if err != nil || u == nil {
			color.Red("Unable to find user `%s`. Please check the username, and re-create the user on ServerAuth.", username)
			return
		}

        // Sleep for 1 second.
        time.Sleep(1 * time.Second)

		if err := viper.ReadInConfig(); err != nil {
        	fmt.Println("Failed to read the config file:", viper.ConfigFileUsed())
        }

		// Load existing accounts from config
		var accounts []Account
		configErr := viper.UnmarshalKey("accounts", &accounts)

		if configErr != nil {
			panic("There was a problem setting up the user account. Please try again or contact ServerAuth for assistance")
		}


		// Remove the account and update the config
		// Loop over accounts and search for the username
		var updatedAccounts []Account
		for _, data := range accounts {
			if data.Username != username {
				updatedAccounts = append(updatedAccounts, Account{data.Username, data.ApiKey})
			}
		}

		// Save the updated accounts list
		viper.Set("accounts", updatedAccounts)

		// Sleep for 1 second to allow the config to be written on slower systems
        time.Sleep(1 * time.Second)

		viper.WriteConfig()
		color.Green("\nThe selected account has been removed from ServerAuth.")
		color.Green("\nThe authorized_keys file has been left in tact to allow you to manually update it.")
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)

	// User flag
	removeCmd.Flags().StringVarP(&username, "username", "u", "", "The username of the system account to add to ServerAuth")
	removeCmd.MarkFlagRequired("username")
}
