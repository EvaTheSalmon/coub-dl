package coub

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

func mux(ctx context.Context, videoPath, audioPath, outPath string, coub Coub) error {
	args := []string{
		"-y",
		"-stream_loop", "-1", "-i", videoPath,
		"-i", audioPath,
		"-map", "0:v", "-map", "1:a",
		"-c:v", "copy", "-c:a", "copy",
		"-shortest",
		"-metadata", "title=" + coub.Title,
		"-metadata", "artist=" + coub.Channel.Title,
		"-metadata", "comment=" + buildComment(coub),
		"-metadata", "creation_time=" + coub.UpdatedAt,
		"-loglevel", "error",
		outPath,
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg: %w", err)
	}

	return nil
}
