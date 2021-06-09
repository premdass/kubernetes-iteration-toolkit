/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/awslabs/kubernetes-iteration-toolkit/pkg/apis/infrastructure/v1alpha1"
	"github.com/awslabs/kubernetes-iteration-toolkit/pkg/awsprovider"
)

type Node struct {
	ID         string
	IPAddress  string
	PrivateDNS string
}

func getMasterInstancesFor(ctx context.Context, clusterName string, ec2api *awsprovider.EC2) ([]*Node, error) {
	instances, err := getInstancesFor(ctx, clusterName, ec2api)
	if err != nil {
		return nil, err
	}
	result := []*Node{}
	for _, instance := range instances {
		if aws.StringValue(instance.State.Name) == "pending" || aws.StringValue(instance.State.Name) == "running" {
			for _, tag := range instance.Tags {
				if aws.StringValue(tag.Key) == "Name" &&
					aws.StringValue(tag.Value) == fmt.Sprintf("%s-%s", clusterName, v1alpha1.MasterInstances) {
					result = append(result, &Node{
						ID:         *instance.InstanceId,
						IPAddress:  *instance.PrivateIpAddress,
						PrivateDNS: *instance.PrivateDnsName,
					})
				}
			}
		}
	}
	return result, nil
}

func getEtcdInstancesFor(ctx context.Context, clusterName string, ec2api *awsprovider.EC2) ([]*Node, error) {
	output, err := getInstancesFor(ctx, clusterName, ec2api)
	if err != nil {
		return nil, err
	}
	nodes := []*Node{}
	for _, instance := range output {
		if aws.StringValue(instance.State.Name) == "pending" || aws.StringValue(instance.State.Name) == "running" {
			for _, tag := range instance.Tags {
				if aws.StringValue(tag.Key) == "Name" &&
					aws.StringValue(tag.Value) == fmt.Sprintf("%s-%s", clusterName, v1alpha1.ETCDInstances) {
					nodes = append(nodes, &Node{
						ID:         *instance.InstanceId,
						IPAddress:  *instance.PrivateIpAddress,
						PrivateDNS: *instance.PrivateDnsName,
					})
				}
			}
		}
	}
	return nodes, nil
}

func getInstancesFor(ctx context.Context, clusterName string, ec2api *awsprovider.EC2) ([]*ec2.Instance, error) {
	output, err := ec2api.DescribeInstancesWithContext(ctx, &ec2.DescribeInstancesInput{
		Filters: ec2FilterFor(clusterName),
	})
	if err != nil {
		return nil, err
	}
	result := []*ec2.Instance{}
	for _, reservation := range output.Reservations {
		result = append(result, reservation.Instances...)
	}
	return result, nil
}
