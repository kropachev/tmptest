package main

import (
    "fmt"
    "regexp"
    "strings"
)

var mdImageRe = regexp.MustCompile(`!\[[^\]]*\]\(([^)\s]+)`)

func firstReadmeImageURL(p Project, content string) string {
    for _, m := range mdImageRe.FindAllStringSubmatch(content, -1) {
        if len(m) < 2 || m[1] == "" {
            continue
        }
        url := m[1]
        if strings.Contains(url, "shields.io") {
            continue
        }
        if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
            owner, name, err := parseRepoURL(p.Repo)
            if err != nil {
                continue
            }
            url = strings.TrimPrefix(url, "./")
            url = fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s",
                owner, name, p.Branch, url)
        }
        return url
    }
    return ""
}