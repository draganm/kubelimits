package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli/v2"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	acv1 "k8s.io/api/core/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	app := cli.App{
		Action: func(c *cli.Context) error {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("while getting user's home dir: %w", err)
			}

			config, err := clientcmd.BuildConfigFromFlags("", filepath.Join(homeDir, ".kube", "config"))
			if err != nil {
				return fmt.Errorf("while building k8s config: %w", err)
			}

			clientset, err := kubernetes.NewForConfig(config)
			if err != nil {
				return fmt.Errorf("while creating client: %w", err)
			}

			ctx := context.Background()
			nsClient := clientset.CoreV1().Namespaces()
			namespaces, err := nsClient.List(ctx, v1.ListOptions{Limit: 400})

			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Name", "CPU", "Mem"})

			for _, n := range namespaces.Items {
				// fmt.Println(n.Name)
				podsClient := clientset.CoreV1().Pods(n.Name)

				pods, err := podsClient.List(ctx, v1.ListOptions{Limit: 400})
				if err != nil {
					return fmt.Errorf("while listing pods in %s: %w", n.Name, err)
				}

				totalCPU, err := resource.ParseQuantity("0")
				if err != nil {
					return fmt.Errorf("while parsing zero quantity: %w", err)
				}

				totalMem, err := resource.ParseQuantity("0")
				if err != nil {
					return fmt.Errorf("while parsing zero quantity: %w", err)
				}

				for _, p := range pods.Items {

					if p.Status.Phase != acv1.PodRunning {
						continue
					}
					for _, c := range p.Spec.Containers {
						requests := c.Resources.Requests

						cpu := requests.Cpu()
						if cpu != nil {
							totalCPU.Add(*cpu)
						}

						mem := requests.Memory()
						if mem != nil {
							totalMem.Add(*mem)
						}

					}
					// for _, c := range p.
				}
				table.Append([]string{
					n.Name,
					fmt.Sprintf("%0.2f", totalCPU.AsApproximateFloat64()),
					fmt.Sprintf("%0.2f", totalMem.AsApproximateFloat64()/(1024*1024)),
				})

			}

			table.Render()

			return nil

		},
	}

	app.RunAndExitOnError()
}
