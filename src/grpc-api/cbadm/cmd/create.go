package cmd

import (
	"fmt"

	"github.com/cloud-barista/cb-mcks/src/grpc-api/cbadm/app"
	"github.com/cloud-barista/cb-mcks/src/utils/lang"
	"github.com/spf13/cobra"
)

type CreateOptions struct {
	*app.Options
}

func (o *CreateOptions) Validate(cmdName string) error {
	o.Namespace = lang.NVL(o.Namespace, app.Config.GetCurrentContext().Namespace)
	if o.Namespace == "" {
		return fmt.Errorf("Namespace is required.")
	}
	if cmdName == "node" {
		if clusterName == "" {
			return fmt.Errorf("ClusterName is required.")
		}
	}
	if o.Data == "" && o.Filename == "" {
		return fmt.Errorf("One of -f Filepath or -d data is required")
	}
	return nil
}

func (o *CreateOptions) ConvertData(cmdName string) error {
	// exute
	out, err := app.GetBody(o)
	if err != nil {
		return err
	} else {
		if cmdName == "cluster" {
			o.Data = `{"namespace":"` + o.Namespace + `" , "ReqInfo": ` + string(out) + `}`
		} else if cmdName == "node" {
			o.Data = `{"namespace":"` + o.Namespace + `" , "cluster":"` + clusterName + `" , "ReqInfo": ` + string(out) + `}`
		} else {
			o.Data = string(out)
		}
	}
	return nil
}

func NewCreateCmd(o *app.Options) *cobra.Command {
	oCreate := &CreateOptions{
		Options: o,
	}

	cmds := &cobra.Command{
		Use:   "create",
		Short: "Create command",
		Long:  "This is a create command",
		Run: func(c *cobra.Command, args []string) {
			c.Help()
		},
	}
	cmdCluster := &cobra.Command{
		Use:   "cluster",
		Short: "Create a cluster",
		Long:  "This is a create command for cluster",
		Run: func(cmd *cobra.Command, args []string) {
			app.ValidateError(cmd, oCreate.Validate(cmd.Name()))
			oCreate.ConvertData(cmd.Name())
			SetupAndRun(cmd, o)
		},
	}
	cmds.AddCommand(cmdCluster)

	cmdNode := &cobra.Command{
		Use:   "node (NAME | --name NAME) --cluster CLUSTER_NAME [options]",
		Short: "Create a node",
		Long:  "This is a create command for node",
		Run: func(cmd *cobra.Command, args []string) {
			app.ValidateError(cmd, oCreate.Validate(cmd.Name()))
			oCreate.ConvertData(cmd.Name())
			SetupAndRun(cmd, o)
		},
	}
	cmdNode.Flags().StringVar(&clusterName, "cluster", "", "Name of cluster")
	cmds.AddCommand(cmdNode)
	/*
		cmdCredential := &cobra.Command{
			Use:   "credential",
			Short: "Create a cloud credential",
			Long:  "This is a create command for credential",
			Run: func(cmd *cobra.Command, args []string) {
				oCreate.ConvertData(cmd.Name())
				SetupAndRun(cmd, o)
			},
		}
		cmds.AddCommand(cmdCredential)
	*/
	return cmds
}
