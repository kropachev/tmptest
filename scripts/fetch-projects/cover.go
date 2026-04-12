package main

import (
    "fmt"
    "io"
    "strings"
)

func fetchCoverImage(p Project, readmeContent string) ([]byte, string, error) {
    data, ext, err := fetchSocialPreview(p)
    if err == nil {
        return data, ext, nil
    }

    imageURL := firstReadmeImageURL(p, readmeContent)
    if imageURL == "" {
        return nil, "", fmt.Errorf("no cover image found")
    }
    return downloadImage(imageURL)
}

func downloadImage(imageURL string) ([]byte, string, error) {
    resp, err := httpClient.Get(imageURL)
    if err != nil {
        return nil, "", err
    }
    defer resp.Body.Close()

    if resp.StatusCode != 200 {
        return nil, "", fmt.Errorf("fetch image %s: %s", imageURL, resp.Status)
    }

    data, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024))
    if err != nil {
        return nil, "", err
    }

    ext := extFromContentType(resp.Header.Get("Content-Type"))
    return data, ext, nil
}

func extFromContentType(ct string) string {
    switch {
    case strings.Contains(ct, "image/png"):
        return ".png"
    case strings.Contains(ct, "image/jpeg"):
        return ".jpg"
    case strings.Contains(ct, "image/gif"):
        return ".gif"
    case strings.Contains(ct, "image/webp"):
        return ".webp"
    default:
        return ".png"
    }
}