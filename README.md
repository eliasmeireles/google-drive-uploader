# Google Drive Uploader CLI

A high-performance CLI tool written in Go to upload files to Google Drive. It supports large files (up to 50GB) and automatic directory management.

## Features

- **Large File Support**: Resumable uploads for reliable transfer of large files.
- **Folder Management**: Automatically handles folder creation if the specified folder does not exist.
- **Secure**: Uses standard OAuth 2.0 flow for authentication.
- **Automation Ready**: Configurable token path for CI/CD or Cron jobs.

## Prerequisites

- Go 1.20 or higher
- A Google Cloud Project with the Drive API enabled.
- an `oauth-client-config.json` (also known as `client_secret.json`) file from your Google Cloud Console.

## Installation

```bash
# Clone the repository
git clone https://github.com/eliasferreira/google-driver-uploader.git
cd google-driver-uploader

# Build the binary
go build -o uploader ./cmd/uploader
```

## Getting Credentials

Before using the uploader, you need to obtain two files: `api-key.json` (OAuth client credentials) and `token.json` (user authorization token).

### Step 1: Create OAuth Client Credentials (`api-key.json`)

1. Go to the [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Enable the **Google Drive API**:
   - Navigate to **APIs & Services** > **Library**
   - Search for "Google Drive API"
   - Click **Enable**
4. Create OAuth 2.0 credentials:
   - Go to **APIs & Services** > **Credentials**
   - Click **Create Credentials** > **OAuth client ID**
   - If prompted, configure the OAuth consent screen first:
     - Choose **External** (or Internal if using Google Workspace)
     - Fill in the required fields (app name, user support email, etc.)
     - Add your email to **Test users** if using External
   - For Application type, select **Desktop app**
   - Give it a name (e.g., "Google Drive Uploader")
   - **IMPORTANT**: Add the following URIs to **Authorized redirect URIs**:
     - `http://localhost:54321/callback`
     - `http://localhost:54322/callback`
     - `http://localhost:54323/callback`
     - `http://localhost:54324/callback`
     - `http://localhost:54325/callback`
     - `http://localhost` (as fallback)
   - Click **Create**
5. Download the credentials:
   - Click the download icon (⬇️) next to your newly created OAuth client
   - Save the file as `api-key.json`

The `api-key.json` file should look like this:

```json
{
  "installed": {
    "client_id": "YOUR_CLIENT_ID.apps.googleusercontent.com",
    "project_id": "your-project-id",
    "auth_uri": "https://accounts.google.com/o/oauth2/auth",
    "token_uri": "https://oauth2.googleapis.com/token",
    "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
    "client_secret": "YOUR_CLIENT_SECRET",
    "redirect_uris": ["http://localhost", "http://localhost:54321/callback"]
  }
}
```

### Step 2: Generate OAuth Token (`token.json`)

Use the `--token-gen` flag to generate the token separately from the upload process. The tool will automatically open your browser and handle the callback.

1. Run the uploader in token generation mode:
   ```bash
   ./uploader --token-gen --client-secret ./api-key.json
   ```

2. Your browser will open automatically. Sign in and authorize the application.

3. After successful authorization, the tool will save `token.json` in the current directory and exit.

The `token.json` file now contains your access token, refresh token, **and client credentials**. This means you only need `token.json` for future uploads!

```json
{
  "access_token": "ya29.a0AfH6SMB...",
  "token_type": "Bearer",
  "refresh_token": "1//0gZ9X...",
  "expiry": "2025-12-24T10:30:00.000Z",
  "client_id": "YOUR_CLIENT_ID...",
  "client_secret": "YOUR_CLIENT_SECRET..."
}
```

> [!IMPORTANT]
> Keep both `api-key.json` and `token.json` secure. Never commit them to version control. The `token.json` allows full access to your Google Drive.

> [!TIP]
> For automated environments (Docker, Kubernetes), generate the `token.json` locally first, then deploy both files as Kubernetes secrets or Docker volumes.

## Usage

### First Run (Authentication)
On the first run, the tool will ask you to visit a URL to authorize the application. After authorization, it will save a `token.json` file.

### Basic Upload
Upload a single file to a specific folder:

```bash
./uploader --root-folder-id "YOUR_FOLDER_ID" path/to/file.zip
```

### With Explicit Client Secret
If `token.json` has credentials embedded, you don't need `--client-secret`. But if you want to provide it explicitly:

```bash
./uploader \
  --client-secret ./api-key.json \
  --root-folder-id "YOUR_FOLDER_ID" \
  path/to/file.zip
```

### Bulk Upload (Directory)
You can upload all files in a directory using the `--workdir` flag.

```bash
./uploader \
  --workdir "./backups" \
  --root-folder-id "ROOT_ID" \
  --delete-on-success
```

### Automation & Default Paths
The tool looks for configuration in default paths, making it ideal for Docker and Kubernetes:
- **Token**: `/etc/google-driver-uploader/token.json` (also checks current directory)
- **Credentials**: `/etc/google-driver-uploader/api-key.json` (Optional: Only needed if you need to regenerate a token)

> [!NOTE]
> For automated environments (Docker, Kubernetes), you only need to provide the `token.json` file. The `api-key.json` is **NOT** required for uploads if you used `--token-gen` to create your token.

### Kubernetes CronJob Example

You can run this tool as a CronJob in Kubernetes to automate your backups. Since `token.json` is self-sufficient, the secret is very simple.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: google-drive-uploader-config
  namespace: default
stringData:
  # The content of your generated token.json
  token.json: |
    {
      "access_token": "<YOUR_ACCESS_TOKEN>",
      "token_type": "Bearer",
      "refresh_token": "<YOUR_REFRESH_TOKEN>",
      "expiry": "<YOUR_EXPIRY_DATE>",
      "client_id": "<YOUR_CLIENT_ID>",
      "client_secret": "<YOUR_CLIENT_SECRET>"
    }
---
apiVersion: batch/v1
kind: CronJob
metadata:
  name: google-drive-backup
spec:
  schedule: "0 4 * * *" # Every day at 4:00 AM
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: uploader
            image: ghcr.io/eliasmeireles/cli/google-driver-uploader:latest
            args:
            - --workdir
            - /backups
            - --root-folder-id
            - "YOUR_FOLDER_ID_HERE"
            - --folder-name
            - "MY_FOLDER_NAME"
            - --smart-organize
            - --delete-on-success
            volumeMounts:
            - name: config
              mountPath: /etc/google-driver-uploader
              readOnly: true
          volumes:
          - name: config
            secret:
              secretName: google-drive-uploader-config
```

> [!NOTE]
> Make sure to create a secret named `google-drive-uploader-config` containing your `token.json`. No other files are needed.

### Docker Usage

You can also run the uploader directly using Docker. This is useful for testing or running in non-Kubernetes environments.

**Prerequisites:**
1. You have a valid `token.json` in `/etc/google-driver-uploader/` on your host machine.
2. You have the files you want to upload in a local directory (e.g., `./data`).

**Run the container:**

```bash
docker run --rm \
  -v /path/to/token.json:/etc/google-driver-uploader/token.json:ro \
  -v /path/to/data:/data \
  ghcr.io/eliasmeireles/cli/google-driver-uploader:latest \
  --workdir /data \
  --root-folder-id "YOUR_FOLDER_ID" \
  --smart-organize \
  --delete-on-success
```

This command:
- Mounts your configuration read-only (`:ro`).
- Mounts your local data directory to `/data` in the container.
- processing all files in `/data`.

### Smart Organization

The `--smart-organize` flag enables automatic folder organization based on filename patterns. Files matching the pattern `[service]_backup_[date]_[time]` will be organized into `Service/Date/` folders.

Example:
  my_database_backup_20251224_084205.tar.gz will be uploaded to `<root fold>/MY_DATABASE/2025-12-24/my_database_backup_20251224_084205.tar.gz`

```bash
./uploader \
  --workdir "./backups" \
  --root-folder-id "ROOT_ID" \
  --smart-organize \
  --delete-on-success
```

### Cleanup Mode

The cleanup feature automatically removes old date-based backup folders based on a retention policy. This is useful for managing backup storage and keeping only recent backups.

**How it works:**
1. Recursively traverses all folders starting from `--root-folder-id`
2. Identifies subdirectories whose names match the date pattern (e.g., `2025-01-15`)
3. For each parent folder containing multiple date-named subfolders:
   - Sorts by date (newest first)
   - Keeps the N most recent folders (specified by `--keep`)
   - Moves older folders to Google Drive trash

**Example:**

```bash
# Keep only the most recent backup in each service folder
./uploader \
  --cleanup \
  --keep 1 \
  --match yyyy-MM-dd \
  --root-folder-id "ROOT_ID"

# Keep the 3 most recent backups
./uploader \
  --cleanup \
  --keep 3 \
  --match yyyyMMdd \
  --root-folder-id "ROOT_ID"
```

**Folder Structure Example:**
```
Root Folder
├── Service A
│   ├── 2025-01-10  ← deleted (older)
│   ├── 2025-01-15  ← deleted (older)
│   └── 2025-01-20  ← kept (most recent)
└── Service B
    ├── 2025-01-12  ← deleted (older)
    └── 2025-01-18  ← kept (most recent)
```

With `--keep 1`, only the most recent date folder in each group is kept. With `--keep 2`, the 2 most recent are kept, and so on.

> [!WARNING]
> Cleanup mode moves folders to trash. While they can be recovered from Google Drive trash, use this feature carefully.

### Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--root-folder-id` | ID of the Google Drive folder to save to. | **Required** |
| `--client-secret` | Path to `api-key.json`. Optional if valid token exists. | `/etc/google-driver-uploader/api-key.json` |
| `--token-path` | Path to the OAuth 2.0 token file. | `token.json` or `/etc/.../token.json` |
| `--workdir` | Path to directory to upload all files from. | - |
| `--smart-organize` | Enable automatic folder organization (`Service/Date/File`). | `false` |
| `--delete-on-success`| Delete local file after successful upload. | `false` |
| `--delete-on-done` | Delete local file after upload attempt (even on failure). | `false` |
| `--folder-name` | Sub-folder name to use/create. | - |
| `--file-name` | Name to save the file as on Drive. | Local filename |
| `--cleanup` | Enable cleanup mode to remove old date-based folders. | `false` |
| `--keep` | Number of most recent date folders to keep (cleanup mode). | `1` |
| `--match` | Date pattern to match folder names (e.g., `yyyy-MM-dd`, `yyyyMMdd`). | `yyyy-MM-dd` |
