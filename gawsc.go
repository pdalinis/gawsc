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
)

func init() {
	client := ec2metadata.New(awsSession)
	// is in aws?
	inAWS = client.Available()

	if inAWS {
		// get region
		region, _ := client.Region()
		awsConfig = aws.NewConfig().WithRegion(region)
		// get instance id

		instanceID, _ = client.GetDynamicData("instance-id")
	}
	Load()
}

func Load() error {
	if !inAWS {
		return errors.New("Cannot load configuration values because we are running inside AWS")
	}

	//TODO: handle errors
	ec2Tags, _ = ec2GetTags(instanceID)

	if asg_name, err := Get("aws:autoscaling:groupName"); err == nil {
		asgTags, _ = asgGetTags(asg_name)
	}

	if stack_name, err := Get("aws:cloudformation:stack-name"); err == nil {
		cftOutputs, _ = cftGetOutputs(stack_name)
	}
	return nil
}

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

func ec2GetTags(resourceId string) (Tag, error) {
	svc := ec2.New(awsSession, awsConfig)
	input := &ec2.DescribeTagsInput{
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name:   aws.String("resource-id"),
				Values: []*string{aws.String(resourceId)},
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

func asgGetTags(resourceId string) (Tag, error) {
	svc := autoscaling.New(awsSession, awsConfig)
	input := &autoscaling.DescribeTagsInput{
		Filters: []*autoscaling.Filter{
			&autoscaling.Filter{
				Name:   aws.String("resource-id"),
				Values: []*string{aws.String(resourceId)},
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

func cftGetOutputs(resourceId string) (Tag, error) {
	svc := cloudformation.New(awsSession, awsConfig)
	input := &cloudformation.DescribeStacksInput{
		StackName: aws.String(resourceId),
	}
	output, err := svc.DescribeStacks(input)

	tags := make(Tag, len(output.Stacks[0].Outputs))

	//TODO error if > 1 returned
	for _, tag := range output.Stacks[0].Outputs {
		tags[*tag.OutputKey] = *tag.OutputValue
	}

	return tags, err
}
