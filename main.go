package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

func main() {
	// Parse command-line flags
	columnsFlag := flag.String("column", "", "Comma-separated list of columns to export (e.g., Repository,Tag)")
	flag.Parse()

	// Define all available columns
	allColumns := []string{
		"ID",
		"Repository",
		"Tag",
		"Created",
		"Size (MB)",
		"SharedSize (MB)",
		"VirtualSize (MB)",
		"Containers",
		"Labels",
	}

	// Determine which columns to export
	var selectedColumns []string
	var columnIndices []int
	if *columnsFlag != "" {
		selectedColumns = strings.Split(*columnsFlag, ",")
		// Validate and get indices
		for _, col := range selectedColumns {
			col = strings.TrimSpace(col)
			found := false
			for idx, availCol := range allColumns {
				if col == availCol {
					columnIndices = append(columnIndices, idx)
					found = true
					break
				}
			}
			if !found {
				log.Fatalf("Invalid column name: %s\nAvailable columns: %s", col, strings.Join(allColumns, ", "))
			}
		}
	} else {
		selectedColumns = allColumns
		for i := range allColumns {
			columnIndices = append(columnIndices, i)
		}
	}

	// Create Docker client
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("Failed to create Docker client: %v", err)
	}
	defer cli.Close()

	// List all images
	images, err := cli.ImageList(ctx, image.ListOptions{All: true})
	if err != nil {
		log.Fatalf("Failed to list images: %v", err)
	}

	// Generate filename
	filename := "docker_images.csv"
	if *columnsFlag != "" {
		filename = "docker_images_" + strings.ReplaceAll(*columnsFlag, ",", "_") + ".csv"
	}

	// Create CSV file
	file, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Failed to create CSV file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write CSV header (only selected columns)
	if err := writer.Write(selectedColumns); err != nil {
		log.Fatalf("Failed to write header: %v", err)
	}

	// Write image data
	for _, img := range images {
		var repo, tag string
		if len(img.RepoTags) > 0 {
			// Parse repository and tag from RepoTags
			repoTag := img.RepoTags[0]
			repo = repoTag
			tag = ""
			// Split repo:tag
			for i := len(repoTag) - 1; i >= 0; i-- {
				if repoTag[i] == ':' {
					repo = repoTag[:i]
					tag = repoTag[i+1:]
					break
				}
			}
		} else if len(img.RepoDigests) > 0 {
			repo = img.RepoDigests[0]
			tag = "<none>"
		} else {
			repo = "<none>"
			tag = "<none>"
		}

		// Format labels
		labels := ""
		for k, v := range img.Labels {
			if labels != "" {
				labels += "; "
			}
			labels += fmt.Sprintf("%s=%s", k, v)
		}

		// Build full row with all data
		fullRow := []string{
			img.ID[7:19], // Short ID
			repo,
			tag,
			fmt.Sprintf("%d", img.Created),
			fmt.Sprintf("%.2f", float64(img.Size)/(1024*1024)),
			fmt.Sprintf("%.2f", float64(img.SharedSize)/(1024*1024)),
			fmt.Sprintf("%.2f", float64(img.VirtualSize)/(1024*1024)),
			fmt.Sprintf("%d", img.Containers),
			labels,
		}

		// Filter row to only include selected columns
		var filteredRow []string
		for _, idx := range columnIndices {
			filteredRow = append(filteredRow, fullRow[idx])
		}

		if err := writer.Write(filteredRow); err != nil {
			log.Fatalf("Failed to write row: %v", err)
		}

		// If image has multiple tags, write additional rows
		for i := 1; i < len(img.RepoTags); i++ {
			repoTag := img.RepoTags[i]
			repo = repoTag
			tag = ""
			for j := len(repoTag) - 1; j >= 0; j-- {
				if repoTag[j] == ':' {
					repo = repoTag[:j]
					tag = repoTag[j+1:]
					break
				}
			}

			fullRow := []string{
				img.ID[7:19],
				repo,
				tag,
				fmt.Sprintf("%d", img.Created),
				fmt.Sprintf("%.2f", float64(img.Size)/(1024*1024)),
				fmt.Sprintf("%.2f", float64(img.SharedSize)/(1024*1024)),
				fmt.Sprintf("%.2f", float64(img.VirtualSize)/(1024*1024)),
				fmt.Sprintf("%d", img.Containers),
				labels,
			}

			var filteredRow []string
			for _, idx := range columnIndices {
				filteredRow = append(filteredRow, fullRow[idx])
			}

			if err := writer.Write(filteredRow); err != nil {
				log.Fatalf("Failed to write row: %v", err)
			}
		}
	}

	fmt.Printf("Successfully exported %d images to %s\n", len(images), filename)
}
