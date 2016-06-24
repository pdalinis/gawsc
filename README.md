# *g*o*awsc*onfiguration reader

When initialized this package will:
1. Read all the EC2 tags into memory
1. Read all the AutoScaling group tags into memory (if any)
1. Read all the Cloudformation Outputs into memory (if any)

Calls into the GetValue method will return the value of the tag in the following order:
1. EC2 Tag
1. ASG Tag
1. CFT Tag

If no tag is found, an error is returned.

##Policy
This package needs an AWS role that has permisions to DescribeInstances, DescribeAutoScalingGroups, and DescribeCFT...

Example Policy:
```
TODO: insert policy
```

##TODO
- Add CFT resources
- Re-read values, throw events when they change
