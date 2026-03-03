package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

// YTDLPInfo structs to parse yt-dlp JSON output
type YTDLPInfo struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Filename    string `json:"_filename"`
}

// VideoJob holds the scraped ID and Title for logging
type VideoJob struct {
	ID    string
	Title string
}

var (
	uploadedMutex sync.Mutex
	uploadedFile  = "uploaded.txt"
)

func main() {
	// 1. Load Environment Variables
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ No .env file found, relying on system environment variables")
	}

	targetChannel := os.Getenv("TARGET_CHANNEL")
	if targetChannel == "" {
		log.Fatal("❌ TARGET_CHANNEL is not set in .env")
	}

	maxWorkers := 3
	fmt.Sscanf(os.Getenv("MAX_WORKERS"), "%d", &maxWorkers)
	privacyStatus := os.Getenv("UPLOAD_PRIVACY")
	if privacyStatus == "" {
		privacyStatus = "private"
	}

	// 2. Setup YouTube API Client
	ctx := context.Background()
	youtubeService := getYouTubeService(ctx)

	// 3. Create Temp Directory for downloads
	os.MkdirAll("temp", os.ModePerm)

	// 4. Fetch all Video IDs and Titles from the target channel
	log.Printf("🔍 Scanning channel: %s...", targetChannel)
	videoJobs := fetchChannelVideoIDs(targetChannel)
	log.Printf("✅ Found %d videos in channel.", len(videoJobs))

	// 5. Load already uploaded videos to prevent duplicates
	uploadedMap := loadUploadedIDs()

	// Filter out already uploaded videos
	var pendingVideos []VideoJob
	for _, job := range videoJobs {
		if !uploadedMap[job.ID] {
			pendingVideos = append(pendingVideos, job)
		}
	}
	log.Printf("🚀 %d videos pending download/upload.", len(pendingVideos))

	if len(pendingVideos) == 0 {
		log.Println("🎉 All videos from this channel have already been uploaded!")
		deleteToken() // Clean up token on early exit
		return
	}

	// 6. Start Goroutine Workers
	jobs := make(chan VideoJob, len(pendingVideos))
	var wg sync.WaitGroup

	for w := 1; w <= maxWorkers; w++ {
		wg.Add(1)
		go worker(w, jobs, youtubeService, privacyStatus, &wg)
	}

	// Send jobs to workers
	for _, job := range pendingVideos {
		jobs <- job
	}
	close(jobs)

	// Wait for all workers to finish
	wg.Wait()
	log.Println("🎉 All processing complete!")

	// 7. Delete token.json after execution finishes
	deleteToken()
}

// worker handles the Download -> Upload -> Delete pipeline
func worker(id int, jobs <-chan VideoJob, service *youtube.Service, privacy string, wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range jobs {
		videoURL := fmt.Sprintf("https://www.youtube.com/watch?v=%s", job.ID)
		log.Printf("[Worker %d] Processing video: %s (%s)", id, job.ID, job.Title)

		// Step A: Download and extract metadata using yt-dlp
		videoMeta, err := downloadVideo(job.ID, videoURL)
		if err != nil {
			log.Printf("❌ [Worker %d] Failed to download %s: %v", id, job.ID, err)
			continue
		}

		// Step B: Upload to YouTube
		log.Printf("⬆️ [Worker %d] Uploading '%s'...", id, videoMeta.Title)
		err = uploadToYouTube(service, videoMeta, privacy)
		if err != nil {
			log.Printf("❌ [Worker %d] Failed to upload %s: %v", id, job.ID, err)
			continue
		}

		// Step C: Cleanup local file
		err = os.Remove(videoMeta.Filename)
		if err != nil {
			log.Printf("⚠️ [Worker %d] Failed to delete temp file %s: %v", id, videoMeta.Filename, err)
		} else {
			log.Printf("🗑️ [Worker %d] Deleted temp file: %s", id, videoMeta.Filename)
		}

		// Step D: Mark as uploaded
		markAsUploaded(job.ID)
		log.Printf("✅ [Worker %d] Successfully finished pipeline for: %s", id, job.ID)
	}
}

// downloadVideo runs yt-dlp to download the video and returns its metadata
func downloadVideo(videoID, url string) (*YTDLPInfo, error) {
	outputTemplate := filepath.Join("temp", fmt.Sprintf("%s.%%(ext)s", videoID))

	cmd := exec.Command("yt-dlp",
		"--dump-json",
		"--no-simulate",
		"--merge-output-format", "mp4", // Force mp4 container if merging
		"-f", "bestvideo[ext=mp4]+bestaudio[ext=m4a]/best[ext=mp4]/best",
		"-o", outputTemplate,
		url,
	)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr // Capture stderr so we can see ffmpeg errors if it fails

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("yt-dlp execution failed: %v\nStderr: %s", err, stderr.String())
	}

	var meta YTDLPInfo
	lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	err = json.Unmarshal([]byte(lines[len(lines)-1]), &meta)
	if err != nil {
		return nil, fmt.Errorf("failed to parse yt-dlp json: %v", err)
	}

	// FIX: Dynamically find the actual downloaded file on disk
	matches, _ := filepath.Glob(filepath.Join("temp", videoID+".*"))
	var actualFile string

	for _, match := range matches {
		if !strings.HasSuffix(match, ".part") && !strings.HasSuffix(match, ".ytdl") {
			actualFile = match
			// Prefer the final merged file over intermediate separated streams (like .f137.mp4)
			if !strings.Contains(filepath.Base(match)[len(videoID):], ".f") {
				break
			}
		}
	}

	if actualFile == "" {
		return nil, fmt.Errorf("file not found on disk after download. (Do you have ffmpeg installed?)\nStderr: %s", stderr.String())
	}

	meta.Filename = actualFile // Overwrite with the verified real file path
	return &meta, nil
}

// uploadToYouTube uploads the local file to your authenticated YouTube channel
func uploadToYouTube(service *youtube.Service, meta *YTDLPInfo, privacy string) error {
	file, err := os.Open(meta.Filename)
	if err != nil {
		return fmt.Errorf("could not open video file: %v", err)
	}
	defer file.Close()

	upload := &youtube.Video{
		Snippet: &youtube.VideoSnippet{
			Title:       meta.Title,
			Description: meta.Description,
			CategoryId:  "22", // 22 = People & Blogs
		},
		Status: &youtube.VideoStatus{
			PrivacyStatus: privacy, // "private", "unlisted", or "public"
		},
	}

	call := service.Videos.Insert([]string{"snippet", "status"}, upload)
	response, err := call.Media(file).Do()
	if err != nil {
		return err
	}

	log.Printf("   -> Uploaded successfully! New Video ID: %s", response.Id)
	return nil
}

// fetchChannelVideoIDs uses yt-dlp to scrape all video IDs and Titles from a channel URL
func fetchChannelVideoIDs(channelURL string) []VideoJob {
	cmd := exec.Command("yt-dlp", "--flat-playlist", "--print", "%(id)s|%(title)s", channelURL)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatalf("Failed to fetch channel videos: %v", err)
	}

	var jobs []VideoJob
	scanner := bufio.NewScanner(&out)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			parts := strings.SplitN(line, "|", 2)
			if len(parts) == 2 {
				jobs = append(jobs, VideoJob{ID: parts[0], Title: parts[1]})
			}
		}
	}
	return jobs
}

// loadUploadedIDs reads the uploaded.txt file to remember what we've already processed
func loadUploadedIDs() map[string]bool {
	uploaded := make(map[string]bool)
	file, err := os.Open(uploadedFile)
	if err != nil {
		if os.IsNotExist(err) {
			return uploaded
		}
		log.Fatalf("Failed to open uploaded.txt: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		id := strings.TrimSpace(scanner.Text())
		if id != "" {
			uploaded[id] = true
		}
	}
	return uploaded
}

// markAsUploaded saves the processed video ID to uploaded.txt
func markAsUploaded(videoID string) {
	uploadedMutex.Lock()
	defer uploadedMutex.Unlock()

	f, err := os.OpenFile(uploadedFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Failed to update uploaded.txt: %v", err)
		return
	}
	defer f.Close()

	if _, err := f.WriteString(videoID + "\n"); err != nil {
		log.Printf("Failed to write to uploaded.txt: %v", err)
	}
}

// Helper to delete token.json
func deleteToken() {
	if err := os.Remove("token.json"); err != nil {
		if !os.IsNotExist(err) {
			log.Printf("⚠️ Failed to delete token.json: %v", err)
		}
	} else {
		log.Println("🧹 Deleted token.json successfully.")
	}
}

// ==========================================
// Google OAuth2 Boilerplate
// ==========================================

func getYouTubeService(ctx context.Context) *youtube.Service {
	b, err := os.ReadFile("client_secret.json")
	if err != nil {
		log.Fatalf("Unable to read client_secret.json (Did you download it from Google Cloud Console?): %v", err)
	}

	config, err := google.ConfigFromJSON(b, youtube.YoutubeUploadScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	client := config.Client(ctx, getAndCacheToken(ctx, config))
	service, err := youtube.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve YouTube client: %v", err)
	}
	return service
}

func getAndCacheToken(ctx context.Context, config *oauth2.Config) *oauth2.Token {
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return tok
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("\n🌐 Go to the following link in your browser:\n\n%v\n\n", authURL)
	fmt.Printf("🔑 Type the authorization code here: ")

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("💾 Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}
