package main

import (
	"github.com/clambin/solaredge-monitor/internal/cmd/cli/export"
	"github.com/clambin/solaredge-monitor/internal/cmd/cli/scrape"
	"github.com/clambin/solaredge-monitor/internal/cmd/cli/web"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log/slog"
	"os"
	"time"
)

var (
	version = "change_me"

	configFile string
	cmd        = cobra.Command{
		Use:   "solaredge-monitor",
		Short: "records solar panel output vs. weather conditions",
	}
)

func main() {
	if err := cmd.Execute(); err != nil {
		slog.Error("failed to start", "err", err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	cmd.Version = version
	cmd.PersistentFlags().StringVar(&configFile, "config", "", "Configuration file")
	cmd.PersistentFlags().Bool("debug", false, "Log debug messages")
	_ = viper.BindPFlag("debug", cmd.PersistentFlags().Lookup("debug"))

	cmd.AddCommand(&web.Cmd, &export.Cmd, &scrape.Cmd)
}

func initConfig() {
	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		viper.AddConfigPath("/etc/solaredge/")
		viper.AddConfigPath("$HOME/.solaredge")
		viper.AddConfigPath(".")
		viper.SetConfigName("config")
	}

	viper.SetDefault("server.addr", ":8080")
	viper.SetDefault("prometheus.addr", ":9090")
	viper.SetDefault("scrape.polling", 5*time.Minute)
	viper.SetDefault("scrape.collection", 15*time.Minute)
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.database", "solar")
	viper.SetDefault("database.username", "solar")
	viper.SetDefault("database.password", "")
	viper.SetDefault("solaredge.token", "")
	viper.SetDefault("tado.username", "")
	viper.SetDefault("tado.password", "")
	viper.SetDefault("tado.secret", "")

	viper.SetEnvPrefix("SOLAREDGE")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		slog.Error("failed to read config file", "err", err)
		os.Exit(1)
	}
}
