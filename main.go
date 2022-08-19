package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}

func inTimeSpan(start, end, check time.Time) bool {
	return check.After(start) && check.Before(end)
}

var articleImagesLastHour = prometheus.NewDesc(
	"article_images_last_hour",
	"Images added in the last hour.",
	[]string{"environment"},
	prometheus.Labels{},
)

type Exporter struct {
	awsRegion, environment, bucketName string
}

func NewExporter(awsRegion string, environment string, bucketName string) *Exporter {
	return &Exporter{
		awsRegion:   awsRegion,
		environment: environment,
		bucketName:  bucketName,
	}
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- articleImagesLastHour
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	log.Println("Running scrape")
	e.ReturnImagesArticles(ch)
}

func (e *Exporter) ReturnImagesArticles(ch chan<- prometheus.Metric) {

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(e.awsRegion)},
	)

	if err != nil {
		exitErrorf("Unable to start s3 session, %v", err)
	}

	loc, _ := time.LoadLocation("UTC")
	timeNow := time.Now().In(loc)
	anHourAgo := timeNow.In(loc).Add(time.Duration(-60) * time.Minute)
	year := timeNow.Year()
	month := int(timeNow.Month()) //need to int this otherwise it defaults to string version e.g April and not 4
	day := timeNow.Day()
	bucketKey := fmt.Sprintf("raw/%v/%v/%v", year, month, day)

	// Create S3 service client
	svc := s3.New(sess)
	params := &s3.ListObjectsInput{
		Bucket: aws.String(e.bucketName),
		Prefix: aws.String(bucketKey),
	}

	var s3objectList []*s3.Object
	err = svc.ListObjectsPages(params, func(p *s3.ListObjectsOutput, last bool) (shouldContinue bool) {
		for _, s3Object := range p.Contents {
			if inTimeSpan(anHourAgo, timeNow, *s3Object.LastModified) {
				s3objectList = append(s3objectList, s3Object)
			}
		}
		return true
	})

	if err != nil {
		exitErrorf("Failed to list S3 objects, %v", err)
	}
	imagesLastHour := len(s3objectList)

	log.Printf("imagesLastHour: %v\n", imagesLastHour)

	ch <- prometheus.MustNewConstMetric(
		articleImagesLastHour,
		prometheus.GaugeValue,
		float64(imagesLastHour),
		e.environment,
	)
}

func main() {
	var awsRegion string = os.Getenv("AWS_REGION")
	var environment string = os.Getenv("ENV")
	var bucketName string = os.Getenv("BUCKET_NAME")

	if awsRegion == "" {
		log.Printf("No AWS region set")
	}

	if environment == "" {
		log.Printf("No Environment set")
	}

	if bucketName == "" {
		log.Printf("No Bucket Name set")
	}

	log.Printf("Bucket Name: %v\n", bucketName)
	log.Printf("Environment: %v\n", environment)
	log.Printf("AWS Region: %v\n", awsRegion)

	exporter := NewExporter(awsRegion, environment, bucketName)
	prometheus.MustRegister(exporter)

	http.Handle("/metrics", promhttp.Handler())

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html>
	<head><title>S3 Image Exporter</title></head>
		<body>
			<h1>S3 Image Exporter</h1>
			<p><a href="/metrics">Metrics</a></p>
		</body>
	</html>`))
	})

	http.ListenAndServe(":20000", nil)
}
