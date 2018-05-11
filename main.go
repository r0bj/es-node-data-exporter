package main

import (
	"net/http"
	"io/ioutil"
	"os"

	// "fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/yaml.v2"
	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	ver string = "0.10"
	logDateLayout string = "2006-01-02 15:04:05"
)

var (
	listenAddress = kingpin.Flag("web.listen-address", "Address to listen on for web interface and telemetry.").Default(":9500").String()
	metricsPath = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
	configFile = kingpin.Flag("config.file", "Path to config file.").Default("/etc/es-node-data-exporter.yaml").String()
)

// Config : struct contains config data from file
type Config struct {
	ElasticsearchClusters []struct {
		Name string `yaml:"name"`
		NodesCount uint `yaml:"nodes_count"`
		DataNodesCount uint `yaml:"data_nodes_count"`
	} `yaml:"elasticsearch_clusters"`
}

var (
	nodesCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "elasticsearch_nodes_target_count",
		Help: "Elasticsearch desirable number of nodes.",
	},
	[]string{"cluster"})
	dataNodesCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "elasticsearch_datanodes_target_count",
		Help: "Elasticsearch desirable number of data nodes.",
	},
	[]string{"cluster"})
)

func init() {
	prometheus.MustRegister(nodesCount)
	prometheus.MustRegister(dataNodesCount)
}

func parseConfig(file string) (Config, error) {
	var config Config

	source, err := ioutil.ReadFile(file)
	if err == nil {
		err = yaml.Unmarshal([]byte(source), &config)
		if err != nil {
			return config, err
		}
	} else {
		return config, err
	}

	return config, nil
}

func startUp(listenAddress, configFile, metricsPath string) {
	config, err := parseConfig(configFile)
	if err != nil {
		log.Fatalf("Cannot parse config file %s: %v", configFile, err)
	}

	for _, cluster := range config.ElasticsearchClusters {
		nodesCount.WithLabelValues(cluster.Name).Set(float64(cluster.NodesCount))
		dataNodesCount.WithLabelValues(cluster.Name).Set(float64(cluster.DataNodesCount))
	}

	http.Handle(metricsPath, prometheus.Handler())
	log.Infof("Starting, version %s, binding to %s", ver, listenAddress)
	http.ListenAndServe(listenAddress, nil)
}

func main() {
	customFormatter := new(log.TextFormatter)
	customFormatter.TimestampFormat = logDateLayout
	log.SetFormatter(customFormatter)
	customFormatter.FullTimestamp = true
	log.SetOutput(os.Stdout)

	kingpin.Version(ver)
	kingpin.Parse()

	startUp(*listenAddress, *configFile, *metricsPath)
}
