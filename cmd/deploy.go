package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers"
	"github.com/spf13/cobra"
)

var (
	DefaultActiveRevisionMode armappcontainers.ActiveRevisionsMode = armappcontainers.ActiveRevisionsModeSingle
	DefaultIngressPort        int32                                = 8080
	DefaultIngressEnabled     bool                                 = true
	TargetEnvironment         string
	ContainerAppName          string
	ImageTag                  string
)

func NewCmdDeploy() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "deploy",
		Short: "Deploy application to Azure Container Apps",
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

			if Config.ActiveRevisionsMode == "" {
				Config.ActiveRevisionsMode = DefaultActiveRevisionMode
			}

			setIngressDefaults()
			setImageTag()
			setSecrets()

			cred, err := azidentity.NewDefaultAzureCredential(nil)
			if err != nil {
				log.Fatalf("Failed to get credentials: %v", err)
			}
			ctx := context.Background()
			client, err := armappcontainers.NewContainerAppsClient(Config.SubscriptionID, cred, nil)
			if err != nil {
				log.Fatalf("Failed to create client: %v", err)
			}
			poller, err := client.BeginCreateOrUpdate(ctx, Config.ResourceGroup, Config.ContainerAppName, armappcontainers.ContainerApp{
				Location: &Config.Location,
				Properties: &armappcontainers.ContainerAppProperties{
					Configuration: &armappcontainers.Configuration{
						ActiveRevisionsMode: to.Ptr(Config.ActiveRevisionsMode),
						Registries:          Config.Registries,
						Dapr:                to.Ptr(Config.Dapr),
						Ingress:             to.Ptr(Config.Ingress),
						Secrets:             Config.Secrets,
					},
					ManagedEnvironmentID: to.Ptr(fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.App/managedEnvironments/%s", Config.SubscriptionID, Config.ResourceGroup, Config.ContainerAppEnvironment)),
					Template: &armappcontainers.Template{
						Containers:     Config.Containers,
						RevisionSuffix: to.Ptr(Config.RevisionSuffix),
						Scale:          to.Ptr(Config.Scaling),
						Volumes:        Config.Volumes,
					},
				},
			}, nil)
			if err != nil {
				log.Fatalf("Failed to finish the request: %v", err)
			}
			res, err := poller.PollUntilDone(ctx, nil)
			if err != nil {
				log.Fatalf("Failed to pull the results: %v", err)
			}
			if Config.Ingress.External != nil {
				githubOutput, exists := os.LookupEnv("GITHUB_OUTPUT")
				if exists {
					setGithubOutput(res, githubOutput)
				}
				fmt.Println(*res.Properties.Configuration.Ingress.Fqdn)
			}
		},
	}

	cmd.Flags().StringVarP(&TargetEnvironment, "environment", "e", "", "Environment where to deploy (required)")
	cmd.Flags().StringVarP(&ImageTag, "tag", "t", "", "Docker image tag")
	cmd.Flags().StringVarP(&ContainerAppName, "container-app-name", "a", "", "Name for Container App. Needed only when creating review environments.")
	cmd.MarkFlagRequired("environment")

	return cmd
}

func setIngressDefaults() {
	if Config.Ingress.TargetPort == nil {
		Config.Ingress.TargetPort = &DefaultIngressPort
	}

	if Config.Ingress.External == nil {
		Config.Ingress.External = &DefaultIngressEnabled
	}
}

func setImageTag() {
	if strings.HasSuffix(*Config.Containers[0].Image, "$tag") {
		if ImageTag != "" {
			*Config.Containers[0].Image = strings.Replace(*Config.Containers[0].Image, "$tag", ImageTag, 1)
		} else {
			*Config.Containers[0].Image = strings.Replace(*Config.Containers[0].Image, "$tag", os.Getenv("GITHUB_SHA"), 1)
		}
	}
}

func setSecrets() {
	var secrets []*armappcontainers.Secret
	for _, container := range Config.Containers {
		for _, env := range container.Env {
			if env.SecretRef != nil {
				value := os.Getenv(*env.Name)
				secrets = append(secrets, &armappcontainers.Secret{Name: env.SecretRef, Value: &value})
			}
		}
	}
	for _, registry := range Config.Registries {
		value := os.Getenv(fmt.Sprintf("%s_REGISTRY_PASSWORD", strings.ToUpper(strings.Replace(*registry.Server, ".", "_", -1))))
		secrets = append(secrets, &armappcontainers.Secret{Name: registry.PasswordSecretRef, Value: &value})
	}
	Config.Secrets = secrets
}

func setGithubOutput(res armappcontainers.ContainerAppsClientCreateOrUpdateResponse, githubOutput string) {
	file, err := os.OpenFile(githubOutput, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err.Error())
	}
	defer file.Close()
	_, err = file.WriteString(fmt.Sprintf("fqdn=%s", *res.Properties.Configuration.Ingress.Fqdn))
	if err != nil {
		log.Println(err.Error())
	}
}
