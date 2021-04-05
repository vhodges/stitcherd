package cmd

import (
	"fmt"
	"os"

	"github.com/google/uuid"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vhodges/stitcherd/stitcher"
)

var (
	// Used for flags.
	cfgFile         string
	hostConfigFiles []string
	listenAddress   string
 	adminHostname   string
	workingDirectory   string
	adminEnabled bool

	rootCmd = &cobra.Command{
		Use:   "stitcherd",
		Short: "Site composition server",
		Long: `Stitcherd uses css selectors to retrieve and replace
elements in an HTML document allowing site architects to compose
their site from a number of different and disparate parts.`,
	}

	serverCmd = &cobra.Command{
		Use:   "serve",
		Short: "Serve the one or more websites",
		Long:  `Server one or more websites configured with a --host.hcl`,
		Run: func(cmd *cobra.Command, args []string) {
			if adminHostname == "" {
				adminHostname = uuid.New().String()
			}
			server := &stitcher.Stitcherd{
				ListenAddress: listenAddress,
				WorkingDirectory: workingDirectory,
				AdminHostName: adminHostname,
				AdminEnabled: adminEnabled,
			}

			server.Init().Run(hostConfigFiles)
		},
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cobra.yaml)")
	rootCmd.PersistentFlags().StringSliceVarP(&hostConfigFiles, "host", "", []string{}, "")

	serverCmd.Flags().StringVar(&listenAddress, "listen", "0.0.0.0:3000", "Address the server should listen on")
	serverCmd.Flags().StringVar(&adminHostname, "adminhost", "", "Hostname to use for the admin end point.  If blank will use a secure UUID")
	serverCmd.Flags().StringVar(&workingDirectory, "workingdir", ".", "Workding directory for site files. Defaults to .")
	serverCmd.Flags().BoolVar(&adminEnabled, "enable-admin", false, "Enable Admin/API enpoint(s)")

	rootCmd.AddCommand(serverCmd)
}

func er(msg interface{}) {
	fmt.Println("Error:", msg)
	os.Exit(1)
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			er(err)
		}

		// Search config in home directory with name ".cobra" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".cobra")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
