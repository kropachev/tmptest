package main

import (
    "fmt"
    "io"
    "net/http"
    "regexp"
    "strings"
)

var ogImageRe = regexp.MustCompile(`<meta\s+property="og:image"\s+content="([^"]+)"`)

func fetchSocialPreview(p Project) ([]byte, string, error) {
    owner, name, err := parseRepoURL(p.Repo)
    if err != nil {
        return nil, "", err
    }

    pageURL := fmt.Sprintf("https://github.com/%s/%s", owner, name)
    req, err := http.NewRequest(http.MethodGet, pageURL, nil)
    if err != nil {
        return nil, "", err
    }
    req.Header.Set("User-Agent", "fetch-projects/1.0")

    resp, err := httpClient.Do(req)
    if err != nil {
        return nil, "", err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, "", fmt.Errorf("fetch %s: %s", pageURL, resp.Status)
    }

    // Meta-теги находятся в <head>, достаточно первых 50 КБ
    head, err := io.ReadAll(io.LimitReader(resp.Body, 50*1024))
    if err != nil {
        return nil, "", err
    }

    m := ogImageRe.FindSubmatch(head)
    if len(m) < 2 {
        return nil, "", fmt.Errorf("og:image not found")
    }

    imageURL := string(m[1])

    // Автосгенерированные превью — не настоящие обложки
    if strings.Contains(imageURL, "opengraph.githubassets.com") {
        return nil, "", fmt.Errorf("no custom social preview")
    }

    return downloadImage(imageURL)
}