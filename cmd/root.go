/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/justsushant/one2n-go-bootcamp/redis-go/redis"
	"github.com/justsushant/one2n-go-bootcamp/redis-go/store/inMemoryStore"
	"github.com/spf13/cobra"
)


// rootCmd represents the base command when called without any subcommands
var rootCmd *cobra.Command

func initRootCmd() {
	rootCmd = &cobra.Command{
		Use:   "redis-go",
		Short: "a primitive redis implementation in go",
		Run: func(cmd *cobra.Command, args []string) {
			reader := bufio.NewReader(os.Stdin)
			ctx := cmd.Context()

			for {
				fmt.Print("> ")
				input, err := reader.ReadString('\n')
				if err != nil {
					fmt.Fprintln(os.Stderr, "Error reading input:", err)
					continue
				}
				input = strings.TrimSpace(input)

				// Break loop if the input is 'exit' or 'quit'
				if input == "exit" || input == "quit" {
					break
				}

				// Execute the command
				args := strings.Split(input, " ")
				if len(args) > 0 {
					rootCmd.SetArgs(args)
					if err := rootCmd.ExecuteContext(ctx); err != nil {
						fmt.Fprintln(os.Stderr, "Error executing command:", err)
					}
				}
			}
		},
	}
}

// SET command
func newSetCmd(server *Server) *cobra.Command {
	return &cobra.Command{
		Use:   "SET",
		Short: "sets the value for a particular key",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			key := args[0]
			value := args[1]
			server.SetAction(key, value)
		},
	}
}

// rootCmd and other commands initialization
func init() {
	initRootCmd()
}


// GET command
func newGetCmd(server *Server) *cobra.Command {
	return &cobra.Command{
		Use:   "GET",
		Short: "gets the value for a particular key",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			key := args[0]
			server.GetAction(key)
		},
	}
}


// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	server := &Server{
		db:  redis.GetNewDB(inMemoryStore.NewInMemoryStore()),
		out: os.Stdout,
	}

	rootCmd.AddCommand(newSetCmd(server), newGetCmd(server))

	ctx := context.WithValue(context.Background(), "server", server)
	rootCmd.SetContext(ctx)

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}

func init() {
	// Define your flags and configuration settings here.
}

func main() {
	Execute()
}
