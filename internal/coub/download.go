package coub

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func (c *Client) Download(ctx context.Context, coub Coub, destDir, name string) (string, bool, error) {
	filename := coub.Permalink
	if name != "" {
		filename = strings.TrimSuffix(name, ".mp4")
	}

	if name == "" && !filenameSafe(filename) {
		return "", false, fmt.Errorf("unsafe permalink %q", filename)
	}
	if name != "" && !userNameSafe(filename) {
		return "", false, fmt.Errorf("unsafe -name %q", filename)
	}

	out := filepath.Join(destDir, filename+".mp4")

	if _, err := os.Stat(out); err == nil {
		return out, true, nil
	}

	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return "", false, fmt.Errorf("creating dir: %w", err)
	}

	video := coub.FileVersions.HTML5.Video
	videoURL := bestURL([]MediaVariant{video.Higher, video.High, video.Med})

	audio := coub.FileVersions.HTML5.Audio
	audioURL := bestURL([]MediaVariant{audio.High, audio.Med})

	tmpVideo := filepath.Join(destDir, filename+"_tmp_video.mp4")
	tmpAudio := filepath.Join(destDir, filename+"_tmp_audio.mp3")
	defer os.Remove(tmpVideo)
	defer os.Remove(tmpAudio)

	if err := c.downloadToFile(ctx, videoURL, tmpVideo); err != nil {
		return "", false, fmt.Errorf("downloading video: %w", err)
	}

	if audioURL == "" {
		return out, false, os.Rename(tmpVideo, out)
	}

	if err := c.downloadToFile(ctx, audioURL, tmpAudio); err != nil {
		return "", false, fmt.Errorf("downloading audio: %w", err)
	}

	if err := mux(ctx, tmpVideo, tmpAudio, out, coub); err != nil {
		os.Remove(out)
		return "", false, fmt.Errorf("muxing: %w", err)
	}

	return out, false, nil
}

func bestURL(variants []MediaVariant) string {
	for _, v := range variants {
		if v.URL != "" {
			return v.URL
		}
	}
	return ""
}

func (c *Client) downloadToFile(ctx context.Context, url, path string) error {
	resp, err := c.doGet(ctx, url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("writing file: %w", err)
	}
	return nil
}

func buildComment(coub Coub) string {
	titles := make([]string, len(coub.Tags))
	for i, t := range coub.Tags {
		titles[i] = t.Title
	}

	return fmt.Sprintf("Author: %s\nLink: https://coub.com/view/%s\nTags: %s",
		coub.Channel.Title, coub.Permalink, strings.Join(titles, ";"))
}

var permalinkPattern = regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9_-]*$`)

func filenameSafe(filename string) bool {
	return permalinkPattern.MatchString(filename)
}

func userNameSafe(name string) bool {
	return name != "" &&
		!strings.HasPrefix(name, "-") &&
		!strings.ContainsAny(name, `/\`) &&
		filepath.IsLocal(name)
}
