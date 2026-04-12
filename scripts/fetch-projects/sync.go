package main

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "os"
    "path/filepath"
    "strings"
)

func syncProject(p Project) error {
    readme, err := fetchReadme(p)
    if err != nil {
        return err
    }

    if p.Title == "" {
        p.Title = p.Slug
    }

    body := stripFirstH1(readme)

    if p.Description == "" {
        desc, err := fetchRepoDescription(p.Repo)
        if err == nil && strings.TrimSpace(desc) != "" {
            p.Description = desc
        } else {
            p.Description = firstNonEmptyLine(body)
        }
    }

    dir := filepath.Join("content", "projects", p.Slug)
    if err := os.MkdirAll(dir, 0o755); err != nil {
        return err
    }

    var coverFile string
    if p.FetchImage == nil || *p.FetchImage {
        if data, ext, err := fetchCoverImage(p, readme); err == nil {
            entries, _ := os.ReadDir(dir)
            for _, e := range entries {
                if strings.HasPrefix(e.Name(), "cover.") {
                    os.Remove(filepath.Join(dir, e.Name()))
                }
            }
            coverFile = "cover" + ext
            if err := os.WriteFile(filepath.Join(dir, coverFile), data, 0o644); err != nil {
                fmt.Fprintf(os.Stderr, "  cover write: %v\n", err)
                coverFile = ""
            }
        }
    }

    path := filepath.Join(dir, "index.md")
    f, err := os.Create(path)
    if err != nil {
        return err
    }
    defer f.Close()

    var imageField string
    if coverFile != "" {
        imageField = fmt.Sprintf("\nimage: \"%s\"", coverFile)
    }

    frontMatter := fmt.Sprintf(`---
title: "%s"
slug: "%s"
description: "%s"%s
tags:
  - projects
repo: "%s"
---

`, escapeYAML(p.Title), escapeYAML(p.Slug), escapeYAML(p.Description), imageField, escapeYAML(p.Repo))

    if _, err := io.WriteString(f, frontMatter); err != nil {
        return err
    }
    if _, err := io.WriteString(f, body); err != nil {
        return err
    }

    fmt.Println("Updated", path)
    return nil
}

func fetchReadme(p Project) (string, error) {
    owner, name, err := parseRepoURL(p.Repo)
    if err != nil {
        return "", err
    }

    rawURL := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/README.md", owner, name, p.Branch)

    resp, err := httpClient.Get(rawURL)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return "", fmt.Errorf("fetch %s: %s", rawURL, resp.Status)
    }

    b, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }
    return string(b), nil
}

type githubRepoResponse struct {
    Description string `json:"description"`
}

func fetchRepoDescription(repoURL string) (string, error) {
    owner, name, err := parseRepoURL(repoURL)
    if err != nil {
        return "", err
    }

    apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, name)
    req, err := http.NewRequest(http.MethodGet, apiURL, nil)
    if err != nil {
        return "", err
    }

    // Токен нужен для приватных репозиториев и при rate limiting.
    // В GitHub Actions доступен через secrets.GITHUB_TOKEN.
    if token := strings.TrimSpace(os.Getenv("GITHUB_TOKEN")); token != "" {
        req.Header.Set("Authorization", "Bearer "+token)
    }
    req.Header.Set("Accept", "application/vnd.github+json")

    resp, err := httpClient.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return "", fmt.Errorf("fetch %s: %s", apiURL, resp.Status)
    }

    b, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }

    var r githubRepoResponse
    if err := json.Unmarshal(b, &r); err != nil {
        return "", err
    }
    return r.Description, nil
}