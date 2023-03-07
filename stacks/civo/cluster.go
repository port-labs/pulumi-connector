package civo

import (
	port_ "github.com/dirien/pulumi-port-labs/sdk/go/port"
	"github.com/pulumi/pulumi-civo/sdk/v2/go/civo"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	"strconv"
)

func KubernetesCluster() pulumi.RunFunc {
	return func(ctx *pulumi.Context) error {

		network, err := civo.NewNetwork(ctx, "civo-network", &civo.NetworkArgs{
			Label: pulumi.Sprintf("network-%s-%s", ctx.Project(), ctx.Stack()),
		})
		if err != nil {
			return err
		}

		firewall, err := civo.NewFirewall(ctx, "civo-firewall", &civo.FirewallArgs{
			NetworkId:          network.ID(),
			CreateDefaultRules: pulumi.Bool(true),
		})

		if err != nil {
			return err
		}
		count, _ := strconv.Atoi(config.Get(ctx, "count"))
		cluster, err := civo.NewKubernetesCluster(ctx, "civo-cluster", &civo.KubernetesClusterArgs{
			NetworkId:   network.ID(),
			FirewallId:  firewall.ID(),
			ClusterType: pulumi.String(config.Get(ctx, "type")),
			Cni:         pulumi.String(config.Get(ctx, "cni")),
			Pools: civo.KubernetesClusterPoolsArgs{
				NodeCount: pulumi.Int(count),
				Size:      pulumi.String(config.Get(ctx, "size")),
			},
		})
		if err != nil {
			return err
		}
		_, err = port_.NewEntity(ctx, "entity", &port_.EntityArgs{
			RunId:      pulumi.String(config.Get(ctx, "run_id")),
			Blueprint:  pulumi.String(config.Get(ctx, "blueprint")),
			Identifier: pulumi.String(config.Get(ctx, "entity_identifier")),
			Title:      pulumi.Sprintf("Kubernetes Cluster %s", cluster.Name),
			Properties: port_.EntityPropertyArray{
				&port_.EntityPropertyArgs{
					Name:  pulumi.String("cluster_name"),
					Value: cluster.Name,
				},
				&port_.EntityPropertyArgs{
					Name:  pulumi.String("type"),
					Value: cluster.ClusterType,
				},
				&port_.EntityPropertyArgs{
					Name:  pulumi.String("cni"),
					Value: cluster.Cni,
				},
				&port_.EntityPropertyArgs{
					Name:  pulumi.String("kconfig"),
					Value: cluster.Kubeconfig,
				},
			},
		})
		if err != nil {
			return err
		}

		return nil
	}
}
