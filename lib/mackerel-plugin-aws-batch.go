package mpawsbatch

import (
	"flag"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/batch"
	mp "github.com/mackerelio/go-mackerel-plugin-helper"
)

// GraphDefinition of AwsBatchPlugin
func (p AwsBatchPlugin) GraphDefinition() map[string](mp.Graphs) {
	graphdef := map[string](mp.Graphs){
		"aws.batch.jobs.#": mp.Graphs{
			Label: "AWS Batch Jobs",
			Unit:  "integer",
			Metrics: [](mp.Metrics){
				mp.Metrics{Name: "SUBMITTED", Label: "SUBMITTED"},
				mp.Metrics{Name: "PENDING", Label: "PENDING"},
				mp.Metrics{Name: "RUNNABLE", Label: "RUNNABLE"},
				mp.Metrics{Name: "STARTING", Label: "STARTING"},
				mp.Metrics{Name: "RUNNING", Label: "RUNNING"},
			},
		},
	}
	return graphdef
}

type AwsBatchPlugin struct {
	AccessKeyID     string
	SecretAccessKey string
	Region          string
	Batch           *batch.Batch
	JobQueue        string
}

// FetchMetrics fetch the metrics
func (p AwsBatchPlugin) FetchMetrics() (map[string]interface{}, error) {
	stat := make(map[string]interface{})
	statuses := []string{"SUBMITTED", "PENDING", "RUNNABLE", "STARTING", "RUNNING"}

	for _, s := range statuses {
		n, err := p.getLastPoint(s)
		if err == nil {
			stat["aws.batch.jobs."+p.JobQueue+"."+s] = n
		}
	}
	return stat, nil
}

func (p *AwsBatchPlugin) prepare() error {
	sess, err := session.NewSession()
	if err != nil {
		return err
	}

	config := aws.NewConfig()
	if p.AccessKeyID != "" && p.SecretAccessKey != "" {
		config = config.WithCredentials(credentials.NewStaticCredentials(p.AccessKeyID, p.SecretAccessKey, ""))
	}
	if p.Region != "" {
		config = config.WithRegion(p.Region)
	}
	p.Batch = batch.New(sess, config)

	return nil
}

func (p AwsBatchPlugin) getLastPoint(status string) (float64, error) {
	input := &batch.ListJobsInput{
		JobQueue:  aws.String(p.JobQueue),
		JobStatus: aws.String(status),
	}

	result, err := p.Batch.ListJobs(input)
	if err != nil {
		return 0.0, err
	}
	return float64(len(result.JobSummaryList)), nil
}

func Do() {
	optAccessKeyID := flag.String("access-key-id", "", "AWS Access Key ID")
	optSecretAccessKey := flag.String("secret-access-key", "", "AWS Secret Access Key")
	optRegion := flag.String("region", "", "AWS Batch Job Region")
	optJobQueue := flag.String("job-queue", "", "AWS Batch Job Queue Name")
	flag.Parse()

	var plugin AwsBatchPlugin

	plugin.AccessKeyID = *optAccessKeyID
	plugin.SecretAccessKey = *optSecretAccessKey
	plugin.Region = *optRegion
	plugin.JobQueue = *optJobQueue

	err := plugin.prepare()
	if err != nil {
		log.Fatalln(err)
	}

	helper := mp.NewMackerelPlugin(plugin)
	helper.Run()
}
