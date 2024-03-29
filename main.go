package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

const Version = "1.2.0"

func usage() {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("LINE Sticker Downloader v%s\n", Version))
	sb.WriteString(fmt.Sprintf("Usage: %s [--static] [--force] [--output=folder] packageID\n\n", os.Args[0]))
	sb.WriteString("positional arguments:\n")
	sb.WriteString("  packgeID    Line Package ID from the store\n")

	sb.WriteString("\noptional arguments:\n")
	sb.WriteString("  --output=FOLDER\tDownload to this folder (default: 'output')\n")
	sb.WriteString("  --static\t\tAlways download static PNGs\n")
	sb.WriteString("  --force\t\tOverride existing stickers\n")

	fmt.Print(sb.String())
}

func fileExists(fileName string) bool {
	if _, err := os.Stat(fileName); err == nil {
		return true
	}
	return false
}

func main() {
	outputFolder := flag.String("output", "output", "Download to this folder")
	static := flag.Bool("static", false, "Always download static PNGs")
	force := flag.Bool("force", false, "Override existing stickers")
	flag.Parse()
	if flag.NArg() != 1 {
		usage()
		os.Exit(0)
	}

	packageId, err := strconv.ParseInt(flag.Arg(0), 10, 64)
	if err != nil {
		usage()
		os.Exit(0)
	}

	log.Printf("LINE Sticker Downloader v%s\n", Version)
	log.Print("Getting LINE sticker pack...")

	httpResp, err := http.Get(fmt.Sprintf(MetaUrl, packageId))
	if err != nil {
		log.Fatal(err)
	}

	if httpResp.StatusCode != 200 {
		log.Fatalf("Sticker pack not found (got HTTP code %d)", httpResp.StatusCode)
	}

	defer httpResp.Body.Close()
	body, err := io.ReadAll(httpResp.Body)

	if err != nil {
		log.Fatal(err)
	}

	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		log.Fatal(err)
	}

	log.Printf("=> Found pack '%s' by '%s'!", response.LocalizedTitle(), response.LocalizedAuthor())

	savePath := filepath.Join(*outputFolder, fmt.Sprintf("LINE_%d", response.PackageID))
	err = os.MkdirAll(savePath, 0770)
	if err != nil {
		log.Fatalf("Could not create directory: %v", err)
	}

	log.Printf("=> Downloading %d stickers...", len(response.Stickers))

	var wg sync.WaitGroup
	for _, sticker := range response.Stickers {
		sticker := sticker
		wg.Add(1)

		go func() {
			defer wg.Done()
			stickerPath := filepath.Join(savePath, sticker.FileName())
			if fileExists(stickerPath) && !*force {
				log.Printf("=> Skipping %s (already exists)", sticker.FileName())
				return
			}

			out, err := os.Create(stickerPath)
			if err != nil {
				log.Printf("Could not create file: %v", err)
				return
			}

			var url string
			if response.HasAnimation && !*static {
				url = sticker.AnimatedDownloadUrl(response.PackageID)
			} else {
				url = sticker.DownloadUrl()
			}

			httpResp, err := http.Get(url)
			if err != nil {
				log.Printf("Could not download sticker %d: %v", sticker.ID, err)
				return
			}

			if httpResp.StatusCode != 200 {
				log.Printf("Status code for sticker %d is not 200, but %d", sticker.ID, httpResp.StatusCode)
				return
			}

			defer httpResp.Body.Close()
			_, err = io.Copy(out, httpResp.Body)
			if err != nil {
				log.Printf("Could not write to file: %v", err)
				return
			}
		}()

	}

	wg.Wait()

	log.Print("Writing json file...")

	infoFile, err := os.Create(filepath.Join(savePath, "info.json"))
	if err != nil {
		log.Fatalf("Could not create info file: %v", err)
	}
	defer infoFile.Close()
	infoJson, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		log.Fatalf("Could not marshal info file: %v", err)
	}
	_, err = infoFile.Write(infoJson)
	if err != nil {
		log.Fatalf("Could not write to info file: %v", err)
	}

	log.Print("Writing info file...")

	infoTxt, err := os.Create(filepath.Join(savePath, "info.txt"))
	if err != nil {
		log.Fatalf("Could not create info text file: %v", err)
	}
	defer infoTxt.Close()

	var sb strings.Builder
	sb.WriteString(
		fmt.Sprintf(
			"LINE Sticker Pack ID %d - '%s' by '%s'\n",
			response.PackageID,
			response.LocalizedTitle(),
			response.LocalizedAuthor(),
		),
	)
	sb.WriteString(
		fmt.Sprintf(
			"JSON URL: %s\n\n",
			fmt.Sprintf(MetaUrl, response.PackageID),
		),
	)
	sb.WriteString(
		fmt.Sprintf(
			"%d stickers:\n",
			len(response.Stickers),
		),
	)
	for _, sticker := range response.Stickers {
		sb.WriteString(
			fmt.Sprintf(
				"%s\n",
				sticker.DownloadUrl(),
			),
		)
	}

	if response.HasAnimation {
		sb.WriteString("\nAnimated stickers:\n")
		for _, sticker := range response.Stickers {
			sb.WriteString(
				fmt.Sprintf(
					"%s\n",
					sticker.AnimatedDownloadUrl(response.PackageID),
				),
			)
		}
	}

	_, err = infoTxt.WriteString(sb.String())
	if err != nil {
		log.Fatalf("Could not write to info text file: %v", err)
	}

	log.Print("DONE!")
}
