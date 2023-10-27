package video

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

// Static variable
var (
	LibretranslateHasStarted = false
)

func Transcribe(objectInfo ObjectInfo, targetLanguage string) (string, error) {
	// Check if environment exists
	if _, err := os.Stat("whisper/venv/bin/activate"); os.IsNotExist(err) {
		return "", errors.New("whisper env not found")
	}

	// Create a directory with video name
	if _, err := os.Stat(objectInfo.TargetPath); os.IsNotExist(err) {
		os.Mkdir(objectInfo.VideoName, 0755)
	}

	fmt.Println("Transcribing... ", objectInfo.AudioPath)
	// Run whisper
	cmd := exec.Command("bash", "-c", `source whisper/venv/bin/activate && whisper `+objectInfo.AudioPath+` `+` --model small`+` --output_dir `+objectInfo.TargetPath+`/`+objectInfo.VideoName)

	// run
	err := cmd.Run()
	if err != nil {
		return "", err
	}

	// Read file from objectInfo.TargetPath as `<uuid>.srt`
	file, err := os.ReadFile(objectInfo.TargetPath + "/" + objectInfo.VideoName + "/" + objectInfo.UUID + ".srt")

	if err != nil {
		return "", err
	}

	// Convert to string
	output := string(file)

	return output, nil
}

// Download whisper command line tool by using pip
func InitAutosub() bool {
	// Check if whisper env exists
	if _, err := os.Stat("whisper/venv/bin/activate"); os.IsNotExist(err) {
		// Check if exists python3
		if _, err := exec.LookPath("python3"); err != nil {
			panic("python3 is not installed")
		} else {
			created, err := createEnv()
			if err != nil {
				panic(err)
			}

			if created {
				fmt.Println("Whisper env created")
				return true
			}

			return false
		}
	} else {
		fmt.Println("Whisper env already exists")

		if LibretranslateHasStarted == false {
			fmt.Println("Starting Libretranslate...")

			// Start libretranslate
			go func() {
				// Start libretranslate
				cmd := exec.Command("bash", "-c", `source whisper/venv/bin/activate && libretranslate`)
				err := cmd.Run()

				if err != nil {
					fmt.Println(err)
				}
			}()
		} else {
			fmt.Println("Libretranslate already started")
		}

		return true
	}
}

func createEnv() (bool, error) {
	// make directory if it doesn't exist as `whisper`
	if _, err := os.Stat("whisper/venv"); os.IsNotExist(err) {
		os.Mkdir("whisper", 0755)
	} else {
		return false, nil
	}

	// create python3 env
	cmd := exec.Command("bash", "-c", `python3 -m venv whisper/venv && source whisper/venv/bin/activate && pip install --upgrade pip && pip install git+https://github.com/openai/whisper.git && pip install libretranslate`)

	output, err := cmd.CombinedOutput()

	if err != nil {
		fmt.Println(string(output))
		return false, errors.New("error while creating env")
	}

	go func() {
		// Start libretranslate
		cmd := exec.Command("bash", "-c", `source whisper/venv/bin/activate && libretranslate`)
		err := cmd.Run()

		if err != nil {
			fmt.Println(err)
		}
	}()

	return true, nil
}

func Translate(subtitlePath string, targetLanguage string, sourceLanguage string) error {
	// Check if environment exists
	if _, err := os.Stat("whisper/venv/bin/activate"); os.IsNotExist(err) {
		return errors.New("whisper env not found")
	}

	// url
	url := "http://localhost:5000/translate_file"
	filePath := subtitlePath

	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer file.Close()

	// HTTP Post
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add file
	part, err := writer.CreateFormFile("file", filePath)
	if err != nil {
		fmt.Println(err)
		return err
	}
	_, err = io.Copy(part, file)

	// other data
	writer.WriteField("source", sourceLanguage)
	writer.WriteField("target", targetLanguage)
	writer.Close()

	// http request
	request, err := http.NewRequest("POST", url, body)
	if err != nil {
		fmt.Println(err)
		return err
	}

	request.Header.Set("Content-Type", writer.FormDataContentType())

	// send request
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer response.Body.Close()

	// get response
	responseBody := new(bytes.Buffer)
	_, err = io.Copy(responseBody, response.Body)
	if err != nil {
		fmt.Println(err)
		return err
	}

	// Parse data as json
	var data map[string]interface{}

	err = json.Unmarshal(responseBody.Bytes(), &data)

	// Get `translatedFileUrl`
	translatedFileUrl := data["translatedFileUrl"].(string)

	// Download url to subtitlePath
	downloadFile(subtitlePath, translatedFileUrl)

	// Replace text
	ReplaceText(subtitlePath)

	return nil
}

func downloadFile(filepath string, url string) (err error) {

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func ReplaceText(subtitlePath string) {
	srtFilePath := strings.Split(subtitlePath, ".")[0] + ".srt"

	file, err := os.Open(srtFilePath)
	if err != nil {
		fmt.Println("File cannot be opened:", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var newText []string
	lineNumber := 0
	replacementTexts, err := readLines(subtitlePath)
	if err != nil {
		fmt.Println("Dosya okuma hatası:", err)
		return
	}

	for scanner.Scan() {
		line := scanner.Text()
		lineNumber++

		if (lineNumber-3)%4 == 0 {
			newText = append(newText, replacementTexts[(lineNumber-3)/4])
		} else {
			newText = append(newText, line)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("An error occurred while reading: ", err)
		return
	}

	// Remove current srt file
	err = os.Remove(srtFilePath)
	if err != nil {
		fmt.Println("File cannot be removed:", err)
		return
	}

	// Dosyayı tekrar yazmak için açın.
	outputFile, err := os.Create(srtFilePath)
	if err != nil {
		fmt.Println("File cannot be created:", err)
		return
	}
	defer outputFile.Close()

	// Yeni metni dosyaya yazın.
	for _, line := range newText {
		fmt.Fprintln(outputFile, line)
	}

	fmt.Println("Replaced text.")
}

func readLines(filename string) ([]string, error) {
	var lines []string

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}
