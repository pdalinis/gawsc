package gawsc

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

type Tag map[string]string

type gawsConfig struct {
	instanceId *string
	awsConfig  *aws.Config
	session    *session.Session
	ec2        Tag
	asg        Tag
	cft        Tag
}

func New(awsConfig *aws.Config, instanceId *string) (*gawsConfig, error) {
	config := &gawsConfig{
		instanceId: instanceId,
		awsConfig:  awsConfig,
	}

	if instanceId == nil {
		if id, err := GetInstanceID(); err != nil {
			return nil, err
		} else {
			config.instanceId = &id
		}
	}

	err := config.Load()

	return config, err
}

//Loads key/values from various sources.
func (c *gawsConfig) Load() error {
	c.ec2, _ = c.Ec2GetTags(*c.instanceId)

	if asg_name, err := c.Get("aws:autoscaling:groupName"); err == nil {
		c.asg, _ = c.AsgGetTags(asg_name)
	}

	if stack_name, err := c.Get("aws:cloudformation:stack-name"); err == nil {
		c.cft, _ = c.CftGetOutputs(stack_name)
	}

	return nil
}

//Gets the value of the specified Key. This first looks on EC2 tags, then Autoscaling Tags, then Cloudformation Output. If no value was found, an error is returned.
func (c *gawsConfig) Get(key string) (string, error) {
	if value, ok := c.ec2[key]; ok {
		return value, nil
	}

	if value, ok := c.asg[key]; ok {
		return value, nil
	}

	if value, ok := c.cft[key]; ok {
		return value, nil
	}

	return "", errors.New("Could not find key " + key)
}

func (c *gawsConfig) ToString() string {
	lines := make([]string, len(c.ec2)+len(c.asg)+len(c.cft))

	appendLines := func(name string, tag Tag) {
		for i := range tag {
			lines = append(lines, fmt.Sprintf("%s:%s:%s", name, i, c.ec2[i]))
		}
	}

	appendLines("EC2", c.ec2)
	appendLines("ASG", c.asg)
	appendLines("CFT", c.cft)

	return strings.Join(lines, "\n")
}

//Gets the Id of the instance that this code is running on.
func GetInstanceID() (string, error) {
	url := "http://169.254.169.254/latest/meta-data/instance-id"
	req, _ := http.NewRequest("GET", url, nil)
	client := http.Client{
		Timeout: time.Millisecond * 100,
	}

	resp, err := client.Do(req)

	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = fmt.Errorf("Code %d returned for url %s", resp.StatusCode, url)
		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return "", err
	}

	return string(body), nil
}
