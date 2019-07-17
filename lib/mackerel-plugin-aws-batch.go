package mpawsbatch

import (
	"flag"
	"fmt"
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
				mp.Metrics{Name: "FAILED", Label: "FAILED"},
				mp.Metrics{Name: "SUCCEEDED", Label: "SUCCEEDED"},
			},
		},
	}
	return graphdef
}

type jobQueueNames []string

// AwsBatchPlugin is a mackerel plugin
type AwsBatchPlugin struct {
	AccessKeyID     string
	SecretAccessKey string
	Region          string
	Batch           *batch.Batch
	JobQueues       jobQueueNames
}

// FetchMetrics fetch the metrics
func (p AwsBatchPlugin) FetchMetrics() (map[string]interface{}, error) {
	stat := make(map[string]interface{})
	statuses := []string{"SUBMITTED", "PENDING", "RUNNABLE", "STARTING", "RUNNING", "FAILED", "SUCCEEDED"}

	for _, name := range p.JobQueues {
		for _, s := range statuses {
			n, err := p.getLastPoint(name, s)
			if err != nil {
				return nil, err
			}
			stat["aws.batch.jobs."+name+"."+s] = n
		}
	}
	return stat, nil
}

func (p *AwsBatchPlugin) prepare() error {
	if len(p.JobQueues) == 0 {
		return fmt.Errorf("Missing job queue names")
	}
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

func (p AwsBatchPlugin) getLastPoint(name string, status string) (float64, error) {
	input := &batch.ListJobsInput{
		JobQueue:  aws.String(name),
		JobStatus: aws.String(status),
	}

	result, err := p.Batch.ListJobs(input)
	if err != nil {
		return 0.0, err
	}
	return float64(len(result.JobSummaryList)), nil
}

func (j *jobQueueNames) String() string {
	return fmt.Sprintf("%v", *j)
}

func (j *jobQueueNames) Set(v string) error {
	*j = append(*j, v)
	return nil
}

// Do the plugin
func Do() {
	var plugin AwsBatchPlugin
	var jqn jobQueueNames

	optAccessKeyID := flag.String("access-key-id", "", "AWS Access Key ID")
	optSecretAccessKey := flag.String("secret-access-key", "", "AWS Secret Access Key")
	optRegion := flag.String("region", "", "AWS Batch Job Region")
	flag.Var(&jqn, "job-queue", "AWS Batch Job Queue Name")
	flag.Parse()

	plugin.AccessKeyID = *optAccessKeyID
	plugin.SecretAccessKey = *optSecretAccessKey
	plugin.Region = *optRegion
	plugin.JobQueues = jqn

	err := plugin.prepare()
	if err != nil {
		log.Fatalln(err)
	}

	helper := mp.NewMackerelPlugin(plugin)
	helper.Run()
}
