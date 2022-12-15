package cmd

import (
	"context"
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers"
	"github.com/spf13/cobra"
)

func NewDeleteCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete application from Azure Container Apps",
		PreRun: func(cmd *cobra.Command, args []string) {
			environment, _ := cmd.Flags().GetString("environment")
			if environment == "review" {
				cmd.MarkFlagRequired("container-app-name")
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			selectedEnvironment := cmd.Flag("environment").Value.String()
			EnvironmentConfigs.GetEnvironment(selectedEnvironment)
			if Config.EnvironmentName == "" {
				log.Fatalf("Couldn't find any matching environments with name %s.", selectedEnvironment)
			}

			// If review environment set user defined container name.
			if Config.EnvironmentName == "review" {
				Config.ContainerAppName = ContainerAppName
			}

			cred, err := azidentity.NewDefaultAzureCredential(nil)
			if err != nil {
				log.Fatalf("failed to get credentials: %v", err)
			}
			ctx := context.Background()
			client, err := armappcontainers.NewContainerAppsClient(Config.SubscriptionID, cred, nil)
			if err != nil {
				log.Fatalf("Failed to create client: %v", err)
			}
			poller, err := client.BeginDelete(ctx, Config.ResourceGroup, Config.ContainerAppName, nil)
			if err != nil {
				log.Fatalf("failed to finish the request: %v", err)
			}
			_, err = poller.PollUntilDone(ctx, nil)
			if err != nil {
				log.Fatalf("failed to pull the result: %v", err)
			}

		},
	}

	cmd.Flags().StringVarP(&TargetEnvironment, "environment", "e", "", "Environment to delete (required)")
	cmd.Flags().StringVarP(&ContainerAppName, "container-app-name", "a", "", "Name for Container App. Needed only when deleting review environments.")
	cmd.MarkFlagRequired("environment")

	return cmd
}
