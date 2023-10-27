package video

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	uuid "github.com/google/uuid"
)

type ObjectInfo struct {
	AudioPath  string
	TargetPath string
	VideoName  string
	UUID       string
}

func ExtractAudio(filePath string, audioPath string, id string) ObjectInfo {
	// Check if `ffmpeg` exists
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		panic("ffmpeg is not installed")
	}

	// Make directory if it doesn't exist `assets`
	if _, err := os.Stat("assets"); os.IsNotExist(err) {
		os.Mkdir("assets", 0755)
	}

	if audioPath == "" {
		// Make directory if it doesn't exist
		audioPath = "assets/audio"
		if _, err := os.Stat(audioPath); os.IsNotExist(err) {
			os.Mkdir(audioPath, 0755)
		}
	}

	if id == "" {
		// Create uuid for audio
		uuid := uuid.New()
		id = uuid.String()
	}

	// ffmpeg -i input.mp4 -map 0:a output.mp3
	cmd := exec.Command("ffmpeg", "-i", filePath, "-map", "0:a", audioPath+"/"+id+".mp3")
	err := cmd.Run()

	if err != nil {
		panic(err)
	}

	// Split filePath by slash
	split := strings.Split(filePath, "/")

	// Get file name
	fileName := split[len(split)-1]

	// Get video name without extension
	videoName := strings.Split(fileName, ".")[0]

	// Get until last split
	newPath := strings.Join(split[:len(split)-1], "/")

	return ObjectInfo{
		AudioPath:  audioPath + "/" + id + ".mp3",
		TargetPath: newPath,
		VideoName:  videoName,
		UUID:       id,
	}
}

func BurnSubtitle(filePath string, subtitlePath string) string {
	// Check if `ffmpeg` exists
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		panic("ffmpeg is not installed")
	}

	// Check if directory `processed` exists
	if _, err := os.Stat("processed"); os.IsNotExist(err) {
		os.Mkdir("processed", 0755)
	}

	videoName := strings.Split(filePath, "/")[len(strings.Split(filePath, "/"))-1]

	outputPath := "processed/" + videoName

	cmd := exec.Command("ffmpeg", "-i", filePath, "-vf", "subtitles="+subtitlePath, outputPath)
	err := cmd.Run()

	if err != nil {
		fmt.Println("error while burning subtitle")
		panic(err)
	}

	return filePath
}
