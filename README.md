# Golang Video/Audio to MP3 Converter Service

This Golang service converts uploaded audio/video files from any format to MP3 and stores them in Firebase Storage. I wrote it to overcome NextJS and Vercel Serverless function’s limitation of FFmpeg conversion. It is designed to handle chunked uploads, process the audio files using `FFmpeg`, and return a publicly accessible Firebase URL for the converted file.

---

## **Features**
- **MP3 Conversion**: Converts uploaded audio files to MP3 format using `ffmpeg`.
- **Firebase Storage Integration**: Uploads the converted files in Firebase Storage with metadata and returns a signed URL for the uploaded file.
- **REST API**: Provides a simple REST API for uploading and converting files.
- **CORS Support**: Allows cross-origin requests from specified domains.

---

## **Prerequisites**
Before running the server, ensure you have the following:

1. **Go**: Install Go from [https://golang.org/dl/](https://golang.org/dl/).
2. **Firebase Project**: Set up a Firebase project and enable Firebase Storage.
3. **Firebase Credentials**: Download the Firebase service account credentials file (`credentials.json`) from the Firebase Console.
4. **FFmpeg**: Install `ffmpeg` on your system. You can download it from [https://ffmpeg.org/download.html](https://ffmpeg.org/download.html).

---

## **Installation**

1. **Clone the Repository**:
   ```bash
   git clone https://github.com/MogboPython/video-audio-mp3-converter.git
   cd video-audio-mp3-converter
   ```

2. **Install Dependencies**:
   ```bash
   go mod tidy
   ```

3. **Set Environment Variables**:
   Create a `.env` file in the root directory with the following the `.env.example`

4. **Place Firebase Credentials**:
   Place the `credentials.json` file (downloaded from Firebase Console) in the root directory of the project.

---

## **Running the Server**

To start the server, run:
```bash
go run main.go
```

The server will start and listen on the specified port (default: `8080`). You should see a log message like:
```
Starting server on port 8080
```

---

## **API Endpoint**

### **Convert and Upload Audio**
- **Endpoint**: `POST /convert`
- **Headers**:
  - `X-User-ID`: Unique identifier for the user.
  - `X-Meeting-ID`: Unique identifier for the meeting.
  - `X-Temp-URL`: Temporary URL (optional).
  - `Content-Type`: `application/octet-stream`.
- **Body**: Binary data of the audio file.
- **Response**:
  ```json
  {
    "url": "https://storage.googleapis.com/your-bucket/users/12345/meetings/67890/audio_20231010_123456.mp3?GoogleAccessId=some-access-id&Expires=1614855080&Signature=some-signature",
    "message": "Conversion successful"
  }
  ```

---

## **Testing**

### **Using cURL**
You can test the `/convert` endpoint using `cURL`:
```bash
curl -X POST http://localhost:8080/convert \
  -H "X-User-ID: 12345" \
  -H "X-Meeting-ID: 67890" \
  -H "X-Temp-URL: http://example.com/temp" \
  -H "Content-Type: application/octet-stream" \
  --data-binary @/path/to/your/audiofile.wav
```

### **Using Postman**
1. Set the request method to `POST`.
2. Set the URL to `http://localhost:8080/convert`.
3. Add the required headers.
4. In the "Body" tab, select "Binary" and upload an audio file (e.g., `.wav`).
5. Send the request.

---


## **TODO**
- [x] Receive file as data streams in request and store temporarily
- [x] Set up FFmpeg and use it to convert file to "mp3”
- [x] Store the MP3 file to Firebase bucket and get the signed URL
- [ ] Add rate limiting
- [ ] Implement polling so conversion can be done in the background
- [ ] Proper documentation

---

## **Deployment**
You can deploy this server to any cloud platform that supports Go, such as:
- **Google Cloud Run**
- **AWS Lambda**
- **Heroku**
- **Fly io**
- **Render**

---

## **Contributing**
Contributions are welcome! Please open an issue or submit a pull request for any improvements or bug fixes.
