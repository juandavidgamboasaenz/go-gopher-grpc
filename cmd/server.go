/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/xerrors"

	pb "github.com/LordCeilan/go-gopher-grpc/pkg/gopher"
	"google.golang.org/grpc"
)

const (
	port         = ":9000"
	KuteGoAPIURL = "127.0.0.1:8080"
)

type Server struct {
	pb.UnimplementedGopherServer
}

type Gopher struct {
	URL string `json: "url"`
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Starts the Schema gRPC server",
	Long: `Instances a server with cobra and create a gRPC connection to the
	Proto file gRPC server.`,
	Run: func(cmd *cobra.Command, args []string) {
		lis, err := net.Listen("tcp", port)
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}

		grpcServer := grpc.NewServer()

		pb.RegisterGopherServer(grpcServer, &Server{})
		log.Printf("GRPC server listening on %v", lis.Addr())

		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	},
}

func (s *Server) GetGopher(ctx context.Context, req *pb.GopherRequest) (*pb.GopherReply, error) {
	res := &pb.GopherReply{}

	if req == nil {
		fmt.Println("request must not be nil")
		return res, xerrors.Errorf("request must not be nil")
	}

	if req.Name == "" {
		fmt.Println("name must not be empty in the request")
		return res, xerrors.Errorf("name must not be empty in the request")
	}

	log.Printf("Received: %v", req.GetName())

	parsedUrl, err := url.Parse("http://" + KuteGoAPIURL + "/gophers?name=" + req.GetName())

	if err != nil {
		log.Fatalf("url incorrect %v or not authorized", parsedUrl)
	}

	// fmt.Printf("%q", parsedUrl)
	// fmt.Printf("%s", parsedUrl.String())

	response, err := http.Get(parsedUrl.String())

	if err != nil {
		log.Fatalf("failed to call KuteGoAPI: %v", err)
	}

	defer response.Body.Close()

	if response.Status == "200 OK" {
		body, err := io.ReadAll(response.Body)

		if err != nil {
			log.Fatalf("failed to read response body: %v", err)
		}

		var data []Gopher
		err = json.Unmarshal(body, &data)
		if err != nil {
			log.Fatalf("failed to unmarshal JSON: %v", err)
		}

		var gophers strings.Builder
		for _, gopher := range data {
			gophers.WriteString(gopher.URL + "\n")
		}

		res.Message = gophers.String()

	} else {
		log.Fatal("Can't get the Gopher :c")
	}
	return res, nil
}

func init() {
	rootCmd.AddCommand(serverCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serverCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serverCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
