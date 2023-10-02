package main

import (
	"fmt"
	"github.com/clambin/solaredge-monitor/internal/analyzer"
	"github.com/clambin/solaredge-monitor/internal/repository"
	"github.com/sjwhitworth/golearn/base"
	"github.com/sjwhitworth/golearn/evaluation"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log/slog"
	"math"
	"math/rand"
	"os"
	"strconv"
	"time"
)

var (
	version    = "change-me"
	configFile string
	cmd        = cobra.Command{
		Use:   "analyzer",
		Short: "predict power output based on recorded statistics",
	}
)

func main() {
	if err := cmd.Execute(); err != nil {
		slog.Error("failed to start", err)
		os.Exit(1)
	}
}

func init() {
	power := cobra.Command{
		Use:   "power",
		Short: "predict power output based on env metrics",
		Run:   predictPower,
	}
	weather := cobra.Command{
		Use:   "weather",
		Short: "predict weather output based on env & power metrics",
		Run:   predictWeather,
	}
	cmd.AddCommand(&power)
	cmd.AddCommand(&weather)

	cobra.OnInitialize(initConfig)
	cmd.Version = version
	cmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "Configuration file")
	cmd.PersistentFlags().BoolP("debug", "d", false, "Log debug messages")
	_ = viper.BindPFlag("debug", cmd.PersistentFlags().Lookup("debug"))
	cmd.PersistentFlags().BoolP("include", "i", false, "Include entries with unknown weather")
	_ = viper.BindPFlag("include", cmd.PersistentFlags().Lookup("include"))
	cmd.PersistentFlags().Float64P("ratio", "r", 0.95, "Percentage of data to use for training")
	_ = viper.BindPFlag("ratio", cmd.PersistentFlags().Lookup("ratio"))
	power.PersistentFlags().Bool("report", false, "Report individual predictions")
	_ = viper.BindPFlag("report", cmd.PersistentFlags().Lookup("report"))
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

	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.database", "solar")
	viper.SetDefault("database.username", "solar")
	viper.SetDefault("database.password", "")

	viper.SetEnvPrefix("SOLAREDGE")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		slog.Error("failed to read config file", err)
		os.Exit(1)
	}
}

func predictPower(_ *cobra.Command, _ []string) {
	var opts slog.HandlerOptions
	if viper.GetBool("debug") {
		opts.Level = slog.LevelDebug
		//opts.AddSource = true
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &opts)))

	slog.Debug("analyzer started", "version", version)

	host := viper.GetString("database.host")
	port := viper.GetInt("database.port")
	database := viper.GetString("database.database")
	username := viper.GetString("database.username")
	password := viper.GetString("database.password")

	db, err := repository.NewPostgresDB(host, port, database, username, password)
	if err != nil {
		slog.Error("failed to access database", err)
		return
	}
	slog.Debug("connected to database", slog.Group("database",
		slog.String("host", host),
		slog.Int("port", port),
		slog.String("database", database),
		slog.String("username", username),
	))

	measurements, err := db.Get(time.Time{}, time.Time{})
	if err != nil {
		slog.Error("failed to get data", err)
		return
	}

	if !viper.GetBool("include") {
		slog.Debug("removing records w/out defined weather")
		measurements = removeUnknown(measurements)
	}

	slog.Debug("records received", "count", len(measurements))

	trainingMeasurements, testMeasurements := splitMeasurements(measurements, viper.GetFloat64("ratio"))
	if len(trainingMeasurements) == 0 {
		slog.Error("no training data. please increase ratio")
		return
	}
	if len(testMeasurements) == 0 {
		slog.Error("no training data. please decrease ratio")
		return
	}
	slog.Debug("records split", "training", len(trainingMeasurements), "test", len(testMeasurements))

	predicted, variance, pctVariance := analyzer.AssessPowerPrediction(trainingMeasurements, testMeasurements)

	if viper.GetBool("report") {
		report(testMeasurements, predicted)
	}

	slog.Info("calculated",
		"delta", strconv.FormatFloat(math.Abs(variance), 'f', 0, 64),
		"variance", strconv.FormatFloat(100*math.Abs(pctVariance), 'f', 0, 64)+"%",
	)
}

func predictWeather(_ *cobra.Command, _ []string) {
	var opts slog.HandlerOptions
	if viper.GetBool("debug") {
		opts.Level = slog.LevelDebug
		//opts.AddSource = true
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &opts)))

	slog.Debug("analyzer started", "version", version)

	host := viper.GetString("database.host")
	port := viper.GetInt("database.port")
	database := viper.GetString("database.database")
	username := viper.GetString("database.username")
	password := viper.GetString("database.password")

	db, err := repository.NewPostgresDB(host, port, database, username, password)
	if err != nil {
		slog.Error("failed to access database", err)
		return
	}
	slog.Debug("connected to database", slog.Group("database",
		slog.String("host", host),
		slog.Int("port", port),
		slog.String("database", database),
		slog.String("username", username),
	))

	measurements, err := db.Get(time.Time{}, time.Time{})
	if err != nil {
		slog.Error("failed to get data", err)
		return
	}

	if !viper.GetBool("include") {
		slog.Debug("removing records w/out defined weather")
		measurements = removeUnknown(measurements)
	}

	allData := analyzer.AnalyzeMeasurements(measurements)
	trainData, testData := base.InstancesTrainTestSplit(allData, .95)
	if r, _ := trainData.Size(); r == 0 {
		slog.Warn("no training data. please increase ratio")
		return
	}
	if r, _ := testData.Size(); r == 0 {
		slog.Warn("no test data. please decrease ratio")
		return
	}

	matrix, err := analyzer.AssessWeatherClassification(trainData, testData)
	if err != nil {
		slog.Error("assess failed", "err", err)
	}

	_, r := testData.Size()
	slog.Info("Accuracy", "testDats", r, "score", strconv.FormatFloat(evaluation.GetAccuracy(matrix), 'f', 3, 64))
	if viper.GetBool("debug") {
		fmt.Println(evaluation.GetSummary(matrix))
	}
}

func removeUnknown(measurements []repository.Measurement) []repository.Measurement {
	var filtered []repository.Measurement
	for _, measurement := range measurements {
		if measurement.Weather != "UNKNOWN" {
			filtered = append(filtered, measurement)
		}
	}
	return filtered
}

func splitMeasurements(measurements []repository.Measurement, ratio float64) ([]repository.Measurement, []repository.Measurement) {
	trainingData := make([]repository.Measurement, 0)
	testData := make([]repository.Measurement, 0)

	for _, measurement := range measurements {
		if rand.Float64() < ratio {
			trainingData = append(trainingData, measurement)
		} else {
			testData = append(testData, measurement)
		}
	}

	return trainingData, testData
}

func report(measurements []repository.Measurement, predict []repository.Measurement) {
	for i, measurement := range measurements {
		fmt.Printf("%s - %3.0f%% %15s measured: %7.0f forecast: %7.0f\n",
			measurement.Timestamp.Format(time.RFC3339),
			measurement.Intensity,
			measurement.Weather,
			measurement.Power,
			predict[i].Power,
		)
	}
}
