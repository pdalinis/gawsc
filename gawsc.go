package gawsc

import (
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// Tag is the base structure for a configuration value
type Tag map[string]string

var (
	inAWS      = false
	awsConfig  *aws.Config
	awsSession = session.New()
	instanceID string
	ec2Tags    Tag
	asgTags    Tag
	cftOutputs Tag
	// Region will contain the AWS region if this code is running in AWS
	Region string
)

func init() {
	client := ec2metadata.New(awsSession)
	// is in aws?
	inAWS = client.Available()

	if inAWS {
		// get region
		Region, _ = client.Region()
		awsConfig = aws.NewConfig().WithRegion(Region)
		// get instance id
		instanceID, _ = client.GetDynamicData("instance-id")
	}
	Load()
}

// Load reads in the configuration values from AWS
func Load() error {
	if !inAWS {
		return errors.New("Cannot load configuration values because we are running inside AWS")
	}

	//TODO: handle errors
	ec2Tags, _ = ec2GetTags(instanceID)

	if asgName, err := Get("aws:autoscaling:groupName"); err == nil {
		asgTags, _ = asgGetTags(asgName)
	}

	if stackName, err := Get("aws:cloudformation:stack-name"); err == nil {
		cftOutputs, _ = cftGetOutputs(stackName)
	}
	return nil
}

// Get a configuration value
func Get(key string) (string, error) {
	if value, ok := ec2Tags[key]; ok {
		return value, nil
	}

	if value, ok := asgTags[key]; ok {
		return value, nil
	}

	if value, ok := cftOutputs[key]; ok {
		return value, nil
	}

	return "", fmt.Errorf("Could not find key %s", key)
}

// GetDefault returns the |defaultValue| if the key is not found.
func GetDefault(key, defaultValue string) (string, error) {
	value, err := Get(key)
	if err != nil {
		return defaultValue, err
	}
	return value, err
}

// ToString outputs all configuration values for debugging purposes
func ToString() string {
	lines := make([]string, len(ec2Tags)+len(asgTags)+len(cftOutputs))

	i := 0
	for k, v := range ec2Tags {
		lines[i] = fmt.Sprintf("ec2Tags:%s%s", k, v)
		i++
	}

	for k, v := range asgTags {
		lines[i] = fmt.Sprintf("asgTags:%s%s", k, v)
		i++
	}

	for k, v := range asgTags {
		lines[i] = fmt.Sprintf("asgTags:%s%s", k, v)
		i++
	}

	return strings.Join(lines, "\n")
}

//TODO: Cleanup code duplication

func ec2GetTags(resourceID string) (Tag, error) {
	svc := ec2.New(awsSession, awsConfig)
	input := &ec2.DescribeTagsInput{
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name:   aws.String("resource-id"),
				Values: []*string{aws.String(resourceID)},
			},
		},
	}
	output, err := svc.DescribeTags(input)

	tags := make(Tag, len(output.Tags))

	//TODO error if > 1 returned
	for _, tag := range output.Tags {
		tags[*tag.Key] = *tag.Value
	}

	return tags, err
}

func asgGetTags(resourceID string) (Tag, error) {
	svc := autoscaling.New(awsSession, awsConfig)
	input := &autoscaling.DescribeTagsInput{
		Filters: []*autoscaling.Filter{
			&autoscaling.Filter{
				Name:   aws.String("resource-id"),
				Values: []*string{aws.String(resourceID)},
			},
		},
	}
	output, err := svc.DescribeTags(input)

	tags := make(Tag, len(output.Tags))

	//TODO error if > 1 returned
	for _, tag := range output.Tags {
		tags[*tag.Key] = *tag.Value
	}

	return tags, err
}

func cftGetOutputs(resourceID string) (Tag, error) {
	svc := cloudformation.New(awsSession, awsConfig)
	input := &cloudformation.DescribeStacksInput{
		StackName: aws.String(resourceID),
	}
	output, err := svc.DescribeStacks(input)

	tags := make(Tag, len(output.Stacks[0].Outputs))

	//TODO error if > 1 returned
	for _, tag := range output.Stacks[0].Outputs {
		tags[*tag.OutputKey] = *tag.OutputValue
	}

	return tags, err
}
