package main

import (
    "bufio"
    "fmt"
    "regexp"
    "strings"
)

var slugRe = regexp.MustCompile(`[^a-zA-Z0-9._-]`)

func parseRepoURL(repoURL string) (owner, name string, err error) {
    trimmed := strings.TrimSuffix(repoURL, "/")
    parts := strings.Split(trimmed, "/")
    if len(parts) < 2 {
        return "", "", fmt.Errorf("invalid repo url: %s", repoURL)
    }
    return parts[len(parts)-2], parts[len(parts)-1], nil
}

func deriveSlugFromRepo(repoURL string) string {
    _, name, err := parseRepoURL(repoURL)
    if err != nil {
        return ""
    }
    return name
}

func escapeYAML(s string) string {
    s = strings.ReplaceAll(s, `\`, `\\`)
    s = strings.ReplaceAll(s, `"`, `\"`)
    return s
}

func sanitizeSlug(s string) string {
    s = strings.ToLower(s)
    s = slugRe.ReplaceAllString(s, "-")
    s = strings.Trim(s, "-")
    if s == "" || s == "." || s == ".." {
        return "_"
    }
    return s
}

func stripFirstH1(content string) string {
    first, rest, found := strings.Cut(content, "\n")
    if strings.HasPrefix(strings.TrimSpace(first), "# ") {
        if found {
            return rest
        }
        return ""
    }
    return content
}

func firstNonEmptyLine(content string) string {
    scanner := bufio.NewScanner(strings.NewReader(content))
    for scanner.Scan() {
        if l := strings.TrimSpace(scanner.Text()); l != "" {
            return l
        }
    }
    return ""
}