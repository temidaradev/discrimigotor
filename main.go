package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"github.com/spf13/cobra"
	"strings"

	"github.com/dhowden/tag"
)

var audioExts = map[string]bool{
	".mp3":  true,
	".flac": true,
	".m4a":  true,
	".wav":  true,
	".ogg":  true,
}

func expandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}

func main() {
	var input string 
	var output string
	cli := &cobra.Command{
		Use: "app",
		Run: func (cmd *cobra.Command, args []string)  {
			fmt.Println("Input: ", input)
			fmt.Println("Output: ", output)
		},
	}
	cli.Flags().StringVar(&input, "input", "", "")
	cli.Flags().StringVar(&output, "output", "", "")

	cli.Execute()

	source := expandHome(input)
	dest := expandHome(output)

	processedAlbums := make(map[string]bool)

	err := filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if !audioExts[ext] {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			fmt.Printf("Error opening %s: %v\n", path, err)
			return nil
		}
		defer file.Close()

		metadata, err := tag.ReadFrom(file)
		if err != nil {
			fmt.Printf("Error reading metadata from %s: %v\n", path, err)
			return nil
		}

		artist := metadata.AlbumArtist()
		if artist == "" {
			artist = metadata.Artist()
		}
		if artist == "" {
			artist = "Unknown Artist"
		}

		album := metadata.Album()
		if album == "" {
			album = "Unknown Album"
		}

		title := metadata.Title()
		if title == "" {
			title = "Unknown Title"
		}

		trackNum, _ := metadata.Track()
		if trackNum == 0 {
			trackNum = 0
		}

		artistDir := filepath.Join(dest, artist)
		albumDir := filepath.Join(artistDir, album)

		err = os.MkdirAll(albumDir, 0755)
		if err != nil {
			fmt.Printf("Error creating directory %s: %v\n", albumDir, err)
			return nil
		}

		albumKey := fmt.Sprintf("%s/%s", artist, album)
		if !processedAlbums[albumKey] {
			processedAlbums[albumKey] = true
			copyCoverArt(filepath.Dir(path), albumDir)
		}

		newName := fmt.Sprintf("%02d - %s%s", trackNum, title, ext)
		destPath := filepath.Join(albumDir, newName)

		if err := copyFile(path, destPath); err != nil {
			fmt.Printf("Error copying %s: %v\n", path, err)
		} else {
			fmt.Printf("Copied: %s\n", destPath)
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error walking source directory: %v\n", err)
	}

	fmt.Println("Done!")
}

func copyCoverArt(sourceDir, destDir string) {
	coverNames := []string{
		"cover.jpg", "cover.png",
		"folder.jpg", "folder.png",
		"album.jpg", "album.png",
		"Cover.jpg", "Cover.png",
		"Folder.jpg", "Folder.png",
		"Album.jpg", "Album.png",
	}

	for _, coverName := range coverNames {
		sourceCover := filepath.Join(sourceDir, coverName)
		if _, err := os.Stat(sourceCover); err == nil {
			destCover := filepath.Join(destDir, strings.ToLower(coverName))
			if err := copyFile(sourceCover, destCover); err == nil {
				fmt.Printf("Copied cover art: %s\n", destCover)
				break
			}
		}
	}
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.Chmod(dst, sourceInfo.Mode())
}
