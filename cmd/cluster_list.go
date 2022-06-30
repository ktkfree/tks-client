/*
Copyright © 2022 SK Telecom <https://github.com/openinfradev>

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
package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"

	"github.com/jedib0t/go-pretty/table"
	pb "github.com/openinfradev/tks-proto/tks_pb"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// clusterListCmd represents the list command
var clusterListCmd = &cobra.Command{
	Use:   "list",
	Short: "Show list of clusters.",
	Long: `Show list of clusters.

Example:
tks cluster list (--long)`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		var conn *grpc.ClientConn
		tksInfoUrl = viper.GetString("tksInfoUrl")
		if tksInfoUrl == "" {
			return errors.New("You must specify tksInfoUrl at config file")
		}
		conn, err := grpc.Dial(tksInfoUrl, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("did not connect: %s", err)
		}
		defer conn.Close()

		client := pb.NewClusterInfoServiceClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()
		data := pb.GetClustersRequest{}
		data.ContractId = viper.GetString("contractId")

		m := protojson.MarshalOptions{
			Indent:        "  ",
			UseProtoNames: true,
		}
		jsonBytes, _ := m.Marshal(&data)
		verbose, err := rootCmd.PersistentFlags().GetBool("verbose")
		long, err := cmd.Flags().GetBool("long")
		if verbose {
			fmt.Println("Proto Json data...")
			fmt.Println(string(jsonBytes))
		}
		r, err := client.GetClusters(ctx, &data)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		} else {
			if len(r.Clusters) == 0 {
				fmt.Println("No cluster exists for specified contract!")
			} else {
				printClusters(filterResponse(r), long)
			}
		}
		return nil
	},
}

func init() {
	clusterCmd.AddCommand(clusterListCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// clusterListCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	clusterListCmd.Flags().BoolP("long", "l", false, "Print detail information")
}

func printClusters(r *pb.GetClustersResponse, long bool) {
	t := table.NewWriter()
	tTemp := table.Table{}
	tTemp.Render()
	t.Style().Options.DrawBorder = false
	t.Style().Options.SeparateColumns = false
	t.Style().Options.SeparateFooter = false
	t.Style().Options.SeparateHeader = false
	t.Style().Options.SeparateRows = false
	if long {
		t.AppendHeader(table.Row{"Name", "ID", "Status", "CREATED_AT", "UPDATED_AT", "CSP_ID", "CONTRACT_ID", "STATUS_DESC"})
		for _, s := range r.Clusters {
			tCreatedAt := parseTime(s.CreatedAt)
			tUpdatedAt := parseTime(s.UpdatedAt)
			t.AppendRow(table.Row{s.Name, s.Id, s.Status, tCreatedAt, tUpdatedAt, s.CspId, s.ContractId, s.StatusDesc})
		}
	} else {
		t.AppendHeader(table.Row{"Name", "ID", "Status", "CREATED_AT", "UPDATED_AT"})
		for _, s := range r.Clusters {
			tCreatedAt := parseTime(s.CreatedAt)
			tUpdatedAt := parseTime(s.UpdatedAt)
			t.AppendRow(table.Row{s.Name, s.Id, s.Status, tCreatedAt, tUpdatedAt})
		}
	}
	fmt.Println(t.Render())

}

func parseTime(t *timestamppb.Timestamp) string {

	return t.AsTime().Format("2006-01-02 15:04:05")
}

func filterResponse(r *pb.GetClustersResponse) *pb.GetClustersResponse {
	clusters := []*pb.Cluster{}
	for _, cluster := range r.Clusters {
		if cluster.GetStatus() != pb.ClusterStatus_DELETED {
			clusters = append(clusters, cluster)
		}
	}

	r.Clusters = clusters
	return r
}
