package cmd

/*
Copyright © 2019 NAME HERE <EMAIL ADDRESS>

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

import (
	"context"
	"fmt"
	productpb "grpc-mongo-crud/proto"

	"github.com/spf13/cobra"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new product",
	Long: `Create a new product on the server through gRPC. 
	
	A blog post requires an Categoty, Title, Price, Quantity.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		category, err := cmd.Flags().GetString("category")
		title, err := cmd.Flags().GetString("title")
		price, err := cmd.Flags().GetString("price")
		quantity, err := cmd.Flags().GetString("quantity")
		if err != nil {
			return err
		}

		product := &productpb.Product{
			Category: category,
			Title:    title,
			Price:    price,
			Quantity: quantity,
		}

		res, err := client.CreateProduct(
			context.TODO(),
			&productpb.CreateProductReq{
				Product: product,
			},
		)
		if err != nil {
			return err
		}

		fmt.Println("Product created: ", res.Product.Id)
		return nil
	},
}

func init() {
	createCmd.Flags().StringP("category", "c", "", "Add category")
	createCmd.Flags().StringP("title", "t", "", "Title for the product")
	createCmd.Flags().StringP("price", "p", "", "Price of the product")
	createCmd.Flags().StringP("quantity", "q", "", "Product quantity")
	createCmd.MarkFlagRequired("category")
	createCmd.MarkFlagRequired("title")
	createCmd.MarkFlagRequired("price")
	createCmd.MarkFlagRequired("quantity")
	rootCmd.AddCommand(createCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
