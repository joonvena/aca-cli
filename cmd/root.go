package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type EnvironmentList struct {
	Environments []Environment `mapstructure:"environments"`
}

type Environment struct {
	EnvironmentName         string                                  `mapstructure:"name"`
	SubscriptionID          string                                  `mapstructure:"subscription_id"`
	ResourceGroup           string                                  `mapstructure:"resource_group"`
	ContainerAppName        string                                  `mapstructure:"container_app_name"`
	ContainerAppEnvironment string                                  `mapstructure:"container_app_environment"`
	Location                string                                  `mapstructure:"location"`
	ActiveRevisionsMode     armappcontainers.ActiveRevisionsMode    `mapstructure:"active_revision_mode"`
	Containers              []*armappcontainers.Container           `mapstructure:"containers"`
	Dapr                    armappcontainers.Dapr                   `mapstructure:"dapr"`
	Ingress                 armappcontainers.Ingress                `mapstructure:"ingress"`
	Identity                armappcontainers.ManagedServiceIdentity `mapstructure:"managed_identity"`
	Registries              []*armappcontainers.RegistryCredentials `mapstructure:"registries"`
	RevisionSuffix          string                                  `mapstructure:"revision_suffix,omitempty"`
	Scaling                 armappcontainers.Scale                  `mapstructure:"scaling"`
	Secrets                 []*armappcontainers.Secret              `mapstructure:"secrets"`
	Tags                    map[string]*string                      `mapstructure:"tags,omitempty"`
	Volumes                 []*armappcontainers.Volume              `mapstructure:"volumes"`
}

var EnvironmentConfigs EnvironmentList
var Config Environment

func NewCmdRoot() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "aca-cli",
		Short: "Tool to deploy containers to Azure Container Apps.",
	}

	viper.SetConfigFile(GetConfigFile())

	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Can't read config: ", err)
		os.Exit(1)
	}

	err := viper.Unmarshal(&EnvironmentConfigs)
	if err != nil {
		fmt.Println("Unmarshaling failed!")
	}

	cmd.AddCommand(NewCmdDeploy(), NewDeleteCmd())

	return cmd
}

func (e *EnvironmentList) GetEnvironment(selectedEnvironment string) Environment {
	for _, environment := range e.Environments {
		if selectedEnvironment == environment.EnvironmentName {
			Config = environment
		}
	}
	return Environment{}
}

func GetConfigFile() string {
	configPath, isDefined := os.LookupEnv("CONFIG_PATH")
	if !isDefined {
		configPath, err := os.Getwd()
		if err != nil {
			log.Fatalf("Couldn't get current working directory %v", err)
		}
		return filepath.Join(configPath, "deploy.yaml")
	}
	githubWorkspace := os.Getenv("GITHUB_WORKSPACE")
	gitlabProjectDir := os.Getenv("CI_PROJECT_DIR")

	if githubWorkspace != "" {
		return filepath.Join(githubWorkspace, "deploy.yaml")
	}
	if gitlabProjectDir != "" {
		return filepath.Join(gitlabProjectDir, "deploy.yaml")
	}
	return filepath.Join(configPath, "deploy.yaml")
}
