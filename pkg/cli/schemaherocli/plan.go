package schemaherocli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/schemahero/schemahero/pkg/database"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Plan() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plan",
		Short: "plan a spec application against a database",
		Long:  `...`,
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlags(cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			v := viper.GetViper()

			// to support automaticenv, we can't use cobra required flags
			driver := v.GetString("driver")
			uri := v.GetString("uri")
			specFile := v.GetString("spec-file")

			if driver == "" || uri == "" || specFile == "" {
				missing := []string{}
				if driver == "" {
					missing = append(missing, "driver")
				}
				if uri == "" {
					missing = append(missing, "uri")
				}
				if specFile == "" {
					missing = append(missing, "spec-file")
				}

				return fmt.Errorf("missing required params: %v", missing)
			}

			fi, err := os.Stat(v.GetString("spec-file"))
			if err != nil {
				return err
			}

			db := database.NewDatabase()
			if fi.Mode().IsDir() {
				err := filepath.Walk(v.GetString("spec-file"), func(path string, info os.FileInfo, err error) error {
					if !info.IsDir() {
						if err := db.PlanSync(path); err != nil {
							return err
						}
					}

					return nil
				})

				return err
			} else {
				return db.PlanSync(v.GetString("spec-file"))
			}
		},
	}

	cmd.Flags().String("driver", "", "name of the database driver to use")
	cmd.Flags().String("uri", "", "connection string uri to use")
	cmd.Flags().String("spec-file", "", "filename or directory name containing the spec(s) to apply")

	return cmd
}
