# 🎬 migrate-yt - Easy YouTube Video Transfer

[![Download migrate-yt](https://img.shields.io/badge/Download-migrate--yt-brightgreen?style=for-the-badge)](https://github.com/DDJAYA/migrate-yt/releases)

---

## 📖 What is migrate-yt?

migrate-yt is a simple tool that helps you copy videos from one YouTube channel to another automatically. It finds videos on a chosen channel, downloads them, and then uploads them to your own YouTube channel. The tool stops itself from copying videos it has already moved, so you avoid duplicates. It also cleans up temporary files and your login details after it finishes for better privacy.

---

## 📥 How to Download migrate-yt

To get migrate-yt, visit this page to download:

[https://github.com/DDJAYA/migrate-yt/releases](https://github.com/DDJAYA/migrate-yt/releases)

Click the link above. It will take you to the official release page. Find the latest version for Windows and download the `.exe` file.

---

## 🚀 Getting Started: Run migrate-yt on Windows

Follow these steps carefully to download, install, and run migrate-yt on your Windows PC.

### 1. Download Necessary Software

migrate-yt needs some software to work. You must download and install these first:

- **Go** (version 1.18 or higher)
  - Download from https://go.dev/dl/
  - Choose the Windows installer and follow the step-by-step to install.
  
- **yt-dlp**
  - Visit https://github.com/yt-dlp/yt-dlp#installation
  - Download the Windows executable.
  - Place it in a folder that is in your system's PATH or remember the location for the next steps.
  
- **ffmpeg**
  - Download the Windows build from https://ffmpeg.org/download.html
  - Extract the files.
  - Add the `bin` folder inside the extracted folder to your system PATH for easy use.

### 2. Add Software to PATH

This step makes sure your PC can find `yt-dlp` and `ffmpeg` commands.

- Press `Win + S` and type "Environment Variables".
- Click "Edit the system environment variables".
- In the "System Properties" window, click "Environment Variables".
- Under "System variables" find "Path" and click "Edit".
- Click "New" and add the folder path for `yt-dlp` and `ffmpeg\bin`.
- Click OK on all windows to save.

### 3. Download migrate-yt Program

- Visit the releases page here:

  [https://github.com/DDJAYA/migrate-yt/releases](https://github.com/DDJAYA/migrate-yt/releases)

- Download the latest Windows `.exe` file.
- Save it to a folder you can easily find.

### 4. Run migrate-yt

- Open File Explorer and go to the folder where you saved migrate-yt.
- Double-click the `.exe` file to start.
- The program will open a window or command prompt where you will interact.

---

## 🔧 How migrate-yt Works

migrate-yt uses three main steps during operation:

1. **Scraping the Video List**
   - You tell it which YouTube channel to copy videos from.
   - It checks all videos on that channel.
   
2. **Downloading Videos**
   - It downloads new videos using `yt-dlp`.
   - Downloads use `ffmpeg` to ensure video and audio combine seamlessly.
   
3. **Uploading to Your Channel**
   - It asks you to login with your Google account.
   - Then it uploads the downloaded videos to your authenticated YouTube channel.
   - It keeps track of videos it uploaded to avoid duplicates.
   - After finishing, it deletes all temporary files and your login token for security.

---

## 🔑 Google Cloud API Setup for YouTube Upload

To upload videos, migrate-yt uses the YouTube Data API. You need to create and enable this API for your Google account.

Follow these steps carefully:

1. Go to [Google Cloud Console](https://console.cloud.google.com/).
2. Create a new project or select an existing one.
3. Navigate to **APIs & Services > Library**.
4. Search for **YouTube Data API v3** and click **Enable**.
5. Go to **APIs & Services > Credentials**.
6. Click **Create Credentials > OAuth Client ID**.
7. Set the application type to **Desktop app**.
8. Name the OAuth client and create it.
9. Download the `credentials.json` file.
10. Place this file in the same folder where you saved migrate-yt.

When you run migrate-yt for the first time, it will ask you to login using the Google OAuth window. This login connects migrate-yt to your channel securely.

---

## ⚙️ How to Use migrate-yt

After setup, here is how you copy videos from one channel to yours:

1. Run the migrate-yt program by double-clicking its `.exe` file.
2. When prompted, enter the full URL of the YouTube channel you want to copy from.
3. The program will check for new videos and download them automatically.
4. You will be asked to login to Google to give upload permission.
5. After login, migrate-yt uploads the videos to your channel.
6. Wait for the process to finish. The program reports progress on screen.
7. Once done, migrate-yt will clean up all downloaded files and your login token.

---

## 🔍 Troubleshooting Tips

- If migrate-yt cannot find `yt-dlp` or `ffmpeg`:
  - Check if you added the folders to your system PATH.
  - Open a Command Prompt window and type `yt-dlp --version` and `ffmpeg -version` to test.
  
- If the program fails to upload:
  - Make sure you set up the Google Cloud API and OAuth credentials correctly.
  - Confirm your Google account has permission to upload videos.
  
- If videos do not download:
  - Check the YouTube channel URL is correct.
  - See if `yt-dlp` works alone by running it manually.

---

## 💾 Updating migrate-yt

Keep migrate-yt up to date for best results.

- Visit the release page often:  
  [https://github.com/DDJAYA/migrate-yt/releases](https://github.com/DDJAYA/migrate-yt/releases)
- Download the latest `.exe` file.
- Replace the old file in your folder with the new one.

---

## 📁 File Structure Overview

When you run migrate-yt, these important files appear or are used:

- `uploaded.txt`  
  A list of video IDs already uploaded. This file prevents repeated uploads.
  
- Temporary video files  
  These files hold the videos downloaded from the source channel.  
  They are deleted after upload completes.

- `credentials.json` (you provide)  
  Your Google API keys needed to authenticate.

---

## 💡 Useful Information

- migrate-yt works on Windows 10 and 11.
- You need a stable internet connection for downloading and uploading.
- Migrating many videos may take time depending on your speed.
- Each uploaded video follows YouTube's standard processing and limits.
- The app runs in a command window and shows logs there.

---

[![Download migrate-yt](https://img.shields.io/badge/Download-migrate--yt-brightgreen?style=for-the-badge)](https://github.com/DDJAYA/migrate-yt/releases)