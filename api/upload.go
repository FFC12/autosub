package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"os"
	"strings"

	tools "autosub/internal"
)

func Upload(c *fiber.Ctx) error {
	file, err := c.FormFile("fileUpload")
	if err != nil {
		return c.JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	target := c.FormValue("target")
	source := c.FormValue("source")

	// lowercase target
	target = strings.ToLower(target)

	// lowercase source
	source = strings.ToLower(source)

	if target == "" {
		target = "tr"
	}

	if source == "" {
		source = "en"
	}

	buffer := make([]byte, file.Size)

	data, err := file.Open()

	if err != nil {
		return c.JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	result, err := data.Read(buffer)
	if err != nil {
		return c.JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	if result != len(buffer) {
		return c.JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	// File format ".mp4" or ".mkv" from filename
	fileExtension := file.Filename[strings.LastIndex(file.Filename, ".")+1:]

	videoId := uuid.New().String()
	tempName := "playlists/" + videoId + "." + fileExtension

	// Check if folder exists
	if _, err := os.Stat("playlists"); os.IsNotExist(err) {
		err = os.Mkdir("playlists", os.ModePerm)
		if err != nil {
			return c.JSON(fiber.Map{
				"message": err.Error(),
			})
		}
	}

	// Save buffer to file in `playlists`
	err = os.WriteFile(tempName, buffer, os.ModePerm)

	if err != nil {
		return c.JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	if err != nil {
		if err != nil {
			return c.JSON(fiber.Map{
				"message": err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	// Get file path from `playlists`
	filePath := "playlists/" + videoId + "." + fileExtension

	objectInfo := tools.ExtractAudio(filePath, "", videoId)

	started := tools.InitAutosub()
	if started {

		_, err := tools.Transcribe(objectInfo, "English")
		if err != nil {
			panic(err)
		}

		subtitlePath := "playlists/" + videoId + "/" + videoId + ".txt"

		err = tools.Translate(subtitlePath, target, source)

		if err != nil {
			panic(err)
		}

		translatedSubtitlePath := "playlists/" + videoId + "/" + videoId + ".srt"

		generatedVideoPath := tools.BurnSubtitle(tempName, translatedSubtitlePath)

		// Remove directory from `playlists`
		err = os.RemoveAll("playlists/" + videoId)
		if err != nil {
			return c.JSON(fiber.Map{
				"message": err.Error(),
			})
		}

		// Remove video from `playlists`
		err = os.Remove(tempName)
		if err != nil {
			return c.JSON(fiber.Map{
				"message": err.Error(),
			})
		}

		// Remove audio `mp3` from `assets/audio`
		err = os.Remove(objectInfo.AudioPath)
		if err != nil {
			return c.JSON(fiber.Map{
				"message": err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"message":       "success",
			"videoUrl":      "/video/" + videoId + "." + fileExtension,
			"generatedFile": generatedVideoPath,
			"key":           videoId,
		})
	} else {
		return c.JSON(fiber.Map{
			"message": "failed",
		})
	}
}
