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

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Find a Product by its ID",
	Long: `Find a product by it's mongoDB Unique identifier.
	
	If no product is found for the ID it will return a 'Not Found' error`,
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := cmd.Flags().GetString("id")
		category, err := cmd.Flags().GetString("category")
		title, err := cmd.Flags().GetString("title")
		price, err := cmd.Flags().GetString("price")
		quantity, err := cmd.Flags().GetString("quantity")

		// Create an UpdateProductRequest
		req := &productpb.UpdateProductReq{
			&productpb.Product{
				Id:       id,
				Category: category,
				Title:    title,
				Price:    price,
				Quantity: quantity,
			},
		}

		res, err := client.UpdateProduct(context.Background(), req)
		if err != nil {
			return err
		}

		fmt.Println(res)
		return nil
	},
}

func init() {
	updateCmd.Flags().StringP("id", "i", "", "The id of the product")
	updateCmd.Flags().StringP("category", "c", "", "Category of the product")
	updateCmd.Flags().StringP("title", "t", "", "The title of the product")
	updateCmd.Flags().StringP("price", "p", "", "The price of the product")
	updateCmd.Flags().StringP("quantity", "q", "", "The quantity of the product")
	updateCmd.MarkFlagRequired("id")
	rootCmd.AddCommand(updateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// updateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// updateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
