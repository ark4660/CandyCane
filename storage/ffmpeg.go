package storage

import (
	"bytes"
	"os/exec"
	"fmt"
	"log"
)

func ExtractAudio(videoBytes []byte) ([]byte, error) {
	cmd := exec.Command("ffmpeg",
    "-probesize", "50M",
    "-analyzeduration", "100M",
    "-err_detect", "ignore_err", // ← ignore demux errors
    "-i", "pipe:0",
    "-vn",
    "-ar", "16000",
    "-ac", "1",
    "-b:a", "32k",
    "-f", "mp3",
    "pipe:1",
	)

    cmd.Stdin = bytes.NewReader(videoBytes)

        var stdout, stderr bytes.Buffer
        cmd.Stdout = &stdout
        cmd.Stderr = &stderr

        if err := cmd.Run(); err != nil {
            return nil, fmt.Errorf("ffmpeg failed: %s", stderr.String())
        }
        log.Printf("ffmpeg output size: %d bytes", stdout.Len())
        log.Printf("ffmpeg stderr: %s", stderr.String())

        if stdout.Len() == 0 {
        	return nil, fmt.Errorf("ffmpeg produced no output\nstderr: %s", stderr.String())
        }

        return stdout.Bytes(), nil
}
