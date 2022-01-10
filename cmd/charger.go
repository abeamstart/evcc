package cmd

import (
	"github.com/evcc-io/evcc/api"
	"github.com/evcc-io/evcc/server"
	"github.com/evcc-io/evcc/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// chargerCmd represents the charger command
var chargerCmd = &cobra.Command{
	Use:               "charger [name]",
	Short:             "Query configured chargers or set parameters",
	PersistentPreRunE: chargerConfig,
	Run:               chargerRun,
}

var chargerSetCmd = &cobra.Command{
	Use:   "set [name]",
	Short: "Set charger parameters",
	Run:   runChargerSet,
}

func init() {
	rootCmd.AddCommand(chargerCmd)
	chargerCmd.AddCommand(chargerSetCmd)
}

func chargerConfig(cmd *cobra.Command, args []string) error {
	util.LogLevel(viper.GetString("log"), viper.GetStringMapString("levels"))
	log.INFO.Printf("evcc %s (%s)", server.Version, server.Commit)

	// load config
	conf, err := loadConfigFile(cfgFile)

	// setup environment
	if err == nil {
		err = configureEnvironment(conf)
	}

	if err == nil {
		err = cp.configureChargers(conf)
	}

	return err
}

func chargerRun(cmd *cobra.Command, args []string) {
	chargers := cp.chargers
	if len(args) == 1 {
		arg := args[0]
		chargers = map[string]api.Charger{arg: cp.Charger(arg)}
	}

	d := dumper{len: len(chargers)}
	for name, v := range chargers {
		d.DumpWithHeader(name, v)
	}
}
