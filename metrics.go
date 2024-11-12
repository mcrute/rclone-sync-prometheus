package main

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rclone/rclone/fs/rc"
)

const namespace = "rclone"

var (
	lastSuccessTime = prometheus.NewDesc(
		prometheus.BuildFQName("", "", "job_last_success_unixtime"),
		"Last time a batch job successfully finished",
		[]string{"instance"}, nil,
	)
	errorCount = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "error_count"),
		"Number of errors encountered by rclone",
		[]string{"instance"}, nil,
	)
	checkCount = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "check_count"),
		"Number of checks performed by rclone",
		[]string{"instance"}, nil,
	)
	checkTotalCount = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "check_total_count"),
		"Total number of checks to be performed by rclone",
		[]string{"instance"}, nil,
	)
	transfersCount = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "transfers_count"),
		"Number of transfers performed by rclone",
		[]string{"instance"}, nil,
	)
	transfersCountTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "transfers_total_count"),
		"Total number of transfers to be performed by rclone",
		[]string{"instance"}, nil,
	)
	deletedDirs = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "deleted_dirs"),
		"Number of directories deleted by rclone",
		[]string{"instance"}, nil,
	)
	deletedFiles = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "deleted_files"),
		"Number of files deleted by rclone",
		[]string{"instance"}, nil,
	)
	renamedFiles = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "renamed_files"),
		"Number of files renamed by rclone",
		[]string{"instance"}, nil,
	)
	elapsedTime = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "elapsed_time"),
		"Elapsed time that rclone has run, in milliseconds",
		[]string{"instance"}, nil,
	)
	transferSpeed = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "transfer_speed"),
		"Transfer speed for rclone, in bytes/second",
		[]string{"instance"}, nil,
	)
	transferBytes = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "transfer_bytes"),
		"Number of bytes transferred by rclone",
		[]string{"instance"}, nil,
	)
	transferBytesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "transfer_total_bytes"),
		"Total number of bytes to be transferred by rclone",
		[]string{"instance"}, nil,
	)
)

type RcloneCollector struct {
	Instance string
	Metrics  rc.Params
}

func (c *RcloneCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- lastSuccessTime
	ch <- errorCount
	ch <- checkCount
	ch <- checkTotalCount
	ch <- transfersCount
	ch <- transfersCountTotal
	ch <- deletedDirs
	ch <- deletedFiles
	ch <- renamedFiles
	ch <- elapsedTime
	ch <- transferSpeed
	ch <- transferBytes
	ch <- transferBytesTotal
}

func (c *RcloneCollector) Collect(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(
		lastSuccessTime, prometheus.GaugeValue, float64(time.Now().UnixNano())/1e9,
		c.Instance,
	)
	ch <- prometheus.MustNewConstMetric(
		errorCount, prometheus.GaugeValue, float64(c.Metrics["errors"].(int64)),
		c.Instance,
	)
	ch <- prometheus.MustNewConstMetric(
		checkCount, prometheus.GaugeValue, float64(c.Metrics["checks"].(int64)),
		c.Instance,
	)
	ch <- prometheus.MustNewConstMetric(
		checkTotalCount, prometheus.GaugeValue, float64(c.Metrics["totalChecks"].(int64)),
		c.Instance,
	)
	ch <- prometheus.MustNewConstMetric(
		transfersCount, prometheus.GaugeValue, float64(c.Metrics["transfers"].(int64)),
		c.Instance,
	)
	ch <- prometheus.MustNewConstMetric(
		transfersCountTotal, prometheus.GaugeValue, float64(c.Metrics["totalTransfers"].(int64)),
		c.Instance,
	)
	ch <- prometheus.MustNewConstMetric(
		deletedDirs, prometheus.GaugeValue, float64(c.Metrics["deletedDirs"].(int64)),
		c.Instance,
	)
	ch <- prometheus.MustNewConstMetric(
		deletedFiles, prometheus.GaugeValue, float64(c.Metrics["deletes"].(int64)),
		c.Instance,
	)
	ch <- prometheus.MustNewConstMetric(
		renamedFiles, prometheus.GaugeValue, float64(c.Metrics["renames"].(int64)),
		c.Instance,
	)
	ch <- prometheus.MustNewConstMetric(
		elapsedTime, prometheus.GaugeValue, c.Metrics["elapsedTime"].(float64)*1000,
		c.Instance,
	)
	ch <- prometheus.MustNewConstMetric(
		transferSpeed, prometheus.GaugeValue, c.Metrics["speed"].(float64),
		c.Instance,
	)
	ch <- prometheus.MustNewConstMetric(
		transferBytes, prometheus.GaugeValue, float64(c.Metrics["bytes"].(int64)),
		c.Instance,
	)
	ch <- prometheus.MustNewConstMetric(
		transferBytesTotal, prometheus.GaugeValue, float64(c.Metrics["totalBytes"].(int64)),
		c.Instance,
	)
}
