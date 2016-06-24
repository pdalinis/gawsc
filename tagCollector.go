package gawsc

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
)

//TODO: Cleanup code duplication

func (c *gawsConfig) Ec2GetTags(resourceId string) (Tag, error) {
	svc := ec2.New(c.session, c.awsConfig)
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

func (c *gawsConfig) AsgGetTags(resourceId string) (Tag, error) {
	svc := autoscaling.New(c.session, c.awsConfig)
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

func (c *gawsConfig) CftGetOutputs(resourceId string) (Tag, error) {
	svc := cloudformation.New(c.session, c.awsConfig)
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
