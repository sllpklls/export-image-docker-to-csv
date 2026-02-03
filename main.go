package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

func main() {
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

	// Create CSV file
	file, err := os.Create("docker_images.csv")
	if err != nil {
		log.Fatalf("Failed to create CSV file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write CSV header
	header := []string{
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
	if err := writer.Write(header); err != nil {
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

		row := []string{
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

		if err := writer.Write(row); err != nil {
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

			row := []string{
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
			if err := writer.Write(row); err != nil {
				log.Fatalf("Failed to write row: %v", err)
			}
		}
	}

	fmt.Printf("Successfully exported %d images to docker_images.csv\n", len(images))
}
