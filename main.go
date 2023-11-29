package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// global variables
var (
	// input file
	inputFile string
	// output directory
	outputDir string
)

// init function
func init() {
	// Set command-line arguments
	flag.StringVar(&inputFile, "input", "", "Path to video file")
	flag.StringVar(&outputDir, "output", "", "Path to output directory")

	// Parse command-line arguments
	flag.Parse()

	// Initialize log file
	logFile, err := os.OpenFile("logs.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Printf("Error opening log file: %v\n", err)
		return
	}
	defer logFile.Close()

	// Use the log package to log messages to both stdout and the log file
	log.SetOutput(io.MultiWriter(os.Stdout, logFile))
}

// main function
func main() {
	// Show help if no arguments are passed
	if inputFile == "" || outputDir == "" {
		showHelp()
		return
	}

	// Create the output directory if it doesn't exist
	err := createOutputDirectory(outputDir)
	if err != nil {
		log.Fatalf("Error creating output directory: %v", err)
	}

	// Get the resolution of the input video
	resolution, err := getVideoResolution(inputFile)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// Check if the resolution is within the specified range
	if !isResolutionValid(resolution) {
		log.Fatal("Invalid resolution")
	}

	// Generate HLS variants
	var variants = map[string]string{
		"1080p": "1920x1080",
		"720p":  "1280x720",
		"480p":  "854x480",
	}

	for variant, scale := range variants {
		outputVariantDir := filepath.Join(outputDir, variant)
		outputPath := filepath.Join(outputVariantDir, fmt.Sprintf("variant_%s.m3u8", variant))
		cmd := exec.Command("ffmpeg", "-i", inputFile, "-vf", fmt.Sprintf("scale=%s", scale), "-c:a", "aac", "-b:a", "192k", "-c:v", "h264", "-b:v", "2M", "-hls_time", "10", "-hls_list_size", "6", "-hls_flags", "delete_segments", outputPath)
		err := cmd.Run()
		if err != nil {
			log.Printf("Error generating HLS variant %s: %v", variant, err)
			log.Fatal("Error generating HLS variant")
		}
	}
}

func isResolutionValid(resolution string) bool {
	// Split the resolution into width and height
	resolutionSplit := strings.Split(resolution, "x")
	widthStr := resolutionSplit[0]
	heightStr := resolutionSplit[1]

	// Convert width and height to integers
	width, errWidth := strconv.Atoi(widthStr)
	height, errHeight := strconv.Atoi(heightStr)

	// Check for conversion errors
	if errWidth != nil || errHeight != nil {
		log.Printf("Error converting resolution to integers: %v", errWidth)
		return false
	}

	// Check if the width and height are within the specified range
	return width >= 1920 && height >= 1080
}

func showHelp() {
	fmt.Printf("Usage: %s -input <input_file> -output <output_directory>\n", os.Args[0])
	flag.PrintDefaults()
}

func createOutputDirectory(outputDir string) error {
	// Check if the directory already exists
	_, err := os.Stat(outputDir)
	if os.IsNotExist(err) {
		// Create the directory if it doesn't exist
		err := os.MkdirAll(outputDir, 0755)
		if err != nil {
			return fmt.Errorf("error creating output directory: %v", err)
		}
	} else if err != nil {
		// Return error if there's an issue checking directory existence
		return fmt.Errorf("error checking output directory: %v", err)
	}

	return nil
}

func getVideoResolution(inputFile string) (string, error) {
	// Run ffprobe command
	cmd := exec.Command("ffprobe", "-v", "error", "-select_streams", "v:0", "-show_entries", "stream=width,height", "-of", "csv=s=x:p=0", inputFile)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return "", fmt.Errorf("error running ffprobe: %v", err)
	}

	// Parse ffprobe output to get resolution
	resolution := strings.TrimSpace(string(output))
	if resolution == "" {
		return "", fmt.Errorf("unable to determine resolution from ffprobe output")
	}

	return resolution, nil
}
