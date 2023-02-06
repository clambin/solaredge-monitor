package main

import (
	"fmt"
	"github.com/clambin/solaredge-monitor/analyzer"
	"github.com/clambin/solaredge-monitor/store"
	"github.com/clambin/solaredge-monitor/version"
	"github.com/sjwhitworth/golearn/base"
	"github.com/sjwhitworth/golearn/evaluation"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/exp/slog"
	"math"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"time"
)

var (
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
		Run:   PredictPower,
	}
	weather := cobra.Command{
		Use:   "weather",
		Short: "predict weather output based on env & power metrics",
		Run:   PredictWeather,
	}
	cmd.AddCommand(&power)
	cmd.AddCommand(&weather)

	cobra.OnInitialize(initConfig)
	cmd.Version = version.BuildVersion
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

func PredictPower(_ *cobra.Command, _ []string) {
	var opts slog.HandlerOptions
	if viper.GetBool("debug") {
		opts.Level = slog.LevelDebug
		//opts.AddSource = true
	}
	slog.SetDefault(slog.New(opts.NewTextHandler(os.Stdout)))

	slog.Debug("analyzer started", "version", version.BuildVersion)

	host := viper.GetString("database.host")
	port := viper.GetInt("database.port")
	database := viper.GetString("database.database")
	username := viper.GetString("database.username")
	password := viper.GetString("database.password")

	db, err := store.NewPostgresDB(host, port, database, username, password)
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

	measurements, err := db.GetAll()
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
		slog.Warn("no training data. please increase ratio")
		return
	}
	if len(testMeasurements) == 0 {
		slog.Warn("no test data. please decrease ratio")
		return
	}
	slog.Debug("records split", "training", len(trainingMeasurements), "test", len(testMeasurements))

	a := analyzer.NewPowerPredictor()
	a.Train(trainingMeasurements)
	slog.Debug("trained")

	predicted := a.PredictSeries(testMeasurements)
	slog.Debug("predicted")

	sort.Slice(predicted, func(i, j int) bool {
		return predicted[i].Timestamp.Before(predicted[j].Timestamp)
	})
	slog.Debug("sorted predicted")

	var variance, pctVariance float64
	for index, measurement := range testMeasurements {
		variance += math.Abs(measurement.Power - predicted[index].Power)
		if measurement.Power > 0 {
			pctVariance += 100 * math.Abs(measurement.Power-predicted[index].Power) / measurement.Power
		}
	}

	if viper.GetBool("report") {
		report(testMeasurements, predicted)
	}

	slog.Info("calculated",
		"delta", strconv.FormatFloat(variance/float64(len(testMeasurements)), 'f', 0, 64),
		"variance", strconv.FormatFloat(pctVariance/float64(len(testMeasurements)), 'f', 0, 64)+"%",
	)
}

func PredictWeather(_ *cobra.Command, _ []string) {
	var opts slog.HandlerOptions
	if viper.GetBool("debug") {
		opts.Level = slog.LevelDebug
		//opts.AddSource = true
	}
	slog.SetDefault(slog.New(opts.NewTextHandler(os.Stdout)))

	slog.Debug("analyzer started", "version", version.BuildVersion)

	host := viper.GetString("database.host")
	port := viper.GetInt("database.port")
	database := viper.GetString("database.database")
	username := viper.GetString("database.username")
	password := viper.GetString("database.password")

	db, err := store.NewPostgresDB(host, port, database, username, password)
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

	measurements, err := db.GetAll()
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

	w := analyzer.NewWeatherClassifier()

	if err = w.Fit(trainData); err != nil {
		slog.Error("training failed", err)
		return
	}

	predictions, err := w.Predict(testData)
	if err != nil {
		slog.Error("predict failed", err)
		return
	}

	confusionMat, err := evaluation.GetConfusionMatrix(testData, predictions)
	if err != nil {
		slog.Error("Unable to get confusion matrix: %s", err)
		return
	}

	fmt.Println(evaluation.GetSummary(confusionMat))
}

func removeUnknown(measurements []store.Measurement) []store.Measurement {
	var filtered []store.Measurement
	for _, measurement := range measurements {
		if measurement.Weather != "UNKNOWN" {
			filtered = append(filtered, measurement)
		}
	}
	return filtered
}

func splitMeasurements(measurements []store.Measurement, ratio float64) ([]store.Measurement, []store.Measurement) {
	trainingData := make([]store.Measurement, 0)
	testData := make([]store.Measurement, 0)

	for _, measurement := range measurements {
		if rand.Float64() < ratio {
			trainingData = append(trainingData, measurement)
		} else {
			testData = append(testData, measurement)
		}
	}

	return trainingData, testData
}

func report(measurements []store.Measurement, predict []store.Measurement) {
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
