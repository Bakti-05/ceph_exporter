package main

import (
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	pvcUsage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "pvc_disk_usage_bytes",
			Help: "Disk usage of the PVC directory",
		},
		[]string{"directory"},
	)
)

func init() {
	prometheus.MustRegister(pvcUsage)
}

func measureDiskUsage(directory string) float64 {
	cmd := exec.Command("du", "-sb", directory)
	out, err := cmd.Output()
	if err != nil {
		log.Printf("Error running du command: %v", err)
		return 0
	}

	output := strings.Fields(string(out))
	if len(output) == 0 {
		return 0
	}

	usage, err := strconv.ParseFloat(output[0], 64)
	if err != nil {
		log.Printf("Error parsing output: %v", err)
		return 0
	}

	return usage
}

func recordMetrics(baseDir string) {
	go func() {
		for {
			err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					log.Printf("Error walking through path %s: %v", path, err)
					return err
				}

				if info.IsDir() {
					usage := measureDiskUsage(path)
					pvcUsage.WithLabelValues(path).Set(usage)
					log.Printf("Directory: %s, Usage: %f bytes", path, usage)
				}

				return nil
			})
			if err != nil {
				log.Printf("Error walking the path %s: %v", baseDir, err)
			}

			time.Sleep(10 * time.Second)
		}
	}()
}

func main() {
	baseDir := os.Getenv("PVC_BASE_DIRECTORY")
	if baseDir == "" {
		log.Fatal("PVC_BASE_DIRECTORY environment variable is not set")
	}

	recordMetrics(baseDir)

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":9128", nil))
}
