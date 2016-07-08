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
	// InAWS will be true if this code is running in AWS
	InAWS      = false
	awsConfig  *aws.Config
	awsSession = session.New()
	// InstanceID is the AWS EC2 instance ID
	InstanceID string
	ec2Tags    Tag
	asgTags    Tag
	cftOutputs Tag
	// Region will contain the AWS region if this code is running in AWS
	Region string
	// LoadError contains any errors that were encountered during init
	LoadError error
	// AZ will contain the AWS AvailabilityZone if this code is running in AWS
	AZ string
)

func init() {
	client := ec2metadata.New(awsSession)
	// is in aws?
	InAWS = client.Available()

	if InAWS {
		// get region
		Region, _ = client.Region()
		awsConfig = aws.NewConfig().WithRegion(Region)
		// get instance id
		InstanceID, _ = client.GetMetadata("instance-id")
		AZ, _ = client.GetMetadata("placement/availability-zone")
	}
	LoadError = Load()
}

  // Load reads in the configuration values from AWS
  func Load() error {
    if !InAWS {
      return errors.New("Cannot load configuration values - not running inside AWS")
    }

    var err error
    ec2Tags, err = ec2GetTags(InstanceID)
    if err != nil {
      return err
    }

    if asgName, err := Get("aws:autoscaling:groupName"); err == nil {
      asgTags, err = asgGetTags(asgName)
      if err != nil {
        return err
      }
    }

    if stackName, err := Get("aws:cloudformation:stack-name"); err == nil {
      cftOutputs, err = cftGetOutputs(stackName)
      if err != nil {
        return err
      }
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
		lines[i] = fmt.Sprintf("ec2Tags:%s:%s", k, v)
		i++
	}

	for k, v := range asgTags {
		lines[i] = fmt.Sprintf("asgTags:%s:%s", k, v)
		i++
	}

	for k, v := range cftOutputs {
		lines[i] = fmt.Sprintf("cftOutputs:%s:%s", k, v)
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
				Name:   aws.String("auto-scaling-group"),
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
