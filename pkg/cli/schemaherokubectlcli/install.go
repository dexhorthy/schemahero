package schemaherokubectlcli

import (
	"fmt"
	"os"

	"github.com/schemahero/schemahero/pkg/installer"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func InstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "install",
		Short:         "install the schemahero operator to the cluster",
		Long:          `...`,
		SilenceErrors: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlags(cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			v := viper.GetViper()

			if v.GetString("extensions-api") != "" {
				if v.GetString("extensions-api") != "v1" && v.GetString("extensions-api") != "v1beta1" {
					fmt.Printf("Unsupported value in extensions-api %q, only v1 and v1beta1 are supported\n", v.GetString("extensions-api"))
					os.Exit(1)
				}
			}
			if v.GetBool("yaml") {
				manifests, err := installer.GenerateOperatorYAML(v.GetString("extensions-api"))
				if err != nil {
					fmt.Printf("Error: %s\n", err.Error())
					return err
				}

				fmt.Printf("%s\n", manifests)
				return nil
			}
			if err := installer.InstallOperator(); err != nil {
				fmt.Printf("Error: %s\n", err.Error())
				return err
			}

			fmt.Println("The SchemaHero operator has been installed to the cluster")
			return nil
		},
	}

	cmd.Flags().Bool("yaml", false, "Is present, don't install the operator, just generate the yaml")
	cmd.Flags().String("extensions-api", "", "version of apiextensions.k8s.io to generate. if unset, will detect best version from kubernetes version")

	return cmd
}
