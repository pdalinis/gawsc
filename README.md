# *g*o*awsc*onfiguration reader

When initialized this package will:

1. Read all the EC2 tags into memory
1. Read all the AutoScaling group tags into memory (if any)
1. Read all the Cloudformation Outputs into memory (if any)

Calls into the Get method will return the value of the first key in the following order:

1. EC2 Tag
1. ASG Tag
1. CFT Output

If no value is found, an error is returned.

##Policy
This package needs an AWS role that has permisions to ec2:DescribeTags, autoscaling:DescribeTags, and cloudformation:DescribeStacks

An example is located in the cloudformation.json file.

##TODO

- Tests
- Add CFT resources
- Re-read values, throw events when they change
- Reduce duplication
