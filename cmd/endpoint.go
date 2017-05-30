package cmd

import (
	"github.com/Frostman/aptomi/pkg/slinga"
	//log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"fmt"
)

var endpointCmd = &cobra.Command{
	Use:   "endpoint",
	Short: "Services endpoints control",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var endpointCmdShow = &cobra.Command{
	Use:   "show",
	Short: "Show endpoints for deployed services",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		// Load the previous usage state
		state := slinga.LoadServiceUsageState()

		endpoints := state.Endpoints()

		for key, keyEndpoints := range endpoints {
			serviceName, contextName, allocationName, componentName := slinga.ParseServiceUsageKey(key)
			fmt.Println("Service:", serviceName, " |  Context:", contextName, " |  Allocation:", allocationName, " |  Component:", componentName)

			for tp, url := range keyEndpoints {
				fmt.Println("	", tp, url)
			}
		}
	},
}

func init() {
	endpointCmd.AddCommand(endpointCmdShow)
	RootCmd.AddCommand(endpointCmd)
}
