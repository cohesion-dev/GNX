package main

import (
    "fmt"
    "io"
    "os"
    "path/filepath"
    "sort"
    "strings"
    "unicode"

    "qiniu-ai-image-generator/gnxaigc"
)

func normalizeCharacterKey(name string) string {
    return strings.ToLower(strings.TrimSpace(name))
}

func collectOrderedFeatures(order []string, registry map[string]gnxaigc.CharacterFeature) []gnxaigc.CharacterFeature {
    features := make([]gnxaigc.CharacterFeature, 0, len(order))
    for _, key := range order {
        if feature, ok := registry[key]; ok {
            features = append(features, feature)
        }
    }
    return features
}

func findCharacterIndex(order []string, key string) int {
    for idx, existing := range order {
        if existing == key {
            return idx
        }
    }
    return -1
}

func copyFile(src, dst string) error {
    if src == "" || dst == "" {
        return fmt.Errorf("copyFile: empty path")
    }
    if src == dst {
        return nil
    }
    in, err := os.Open(src)
    if err != nil {
        return err
    }
    defer in.Close()

    info, err := in.Stat()
    if err != nil {
        return err
    }

    if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
        return err
    }

    out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
    if err != nil {
        return err
    }
    defer out.Close()

    _, err = io.Copy(out, in)
    return err
}

func collectSegmentCharacterKeys(segment gnxaigc.SourceTextSegment, chapterFeatures []gnxaigc.CharacterFeature, seen map[string]struct{}) {
    appendName := func(name string) {
        key := normalizeCharacterKey(name)
        if key == "" {
            return
        }
        seen[key] = struct{}{}
    }

    for _, name := range segment.CharacterNames {
        appendName(name)
    }

    for _, idx := range segment.CharacterRefs {
        if idx < 0 || idx >= len(chapterFeatures) {
            continue
        }
        appendName(chapterFeatures[idx].Basic.Name)
    }
}

func collectPageCharacterKeys(page gnxaigc.StoryboardPage, chapterFeatures []gnxaigc.CharacterFeature) []string {
    seen := make(map[string]struct{})
    for _, panel := range page.Panels {
        for _, segment := range panel.SourceTextSegments {
            collectSegmentCharacterKeys(segment, chapterFeatures, seen)
        }
    }

    keys := make([]string, 0, len(seen))
    for key := range seen {
        keys = append(keys, key)
    }
    sort.Strings(keys)
    return keys
}

func sanitizeCharacterFileStem(name string, index int) string {
    base := strings.TrimSpace(name)
    if base == "" {
        return fmt.Sprintf("character_%02d", index+1)
    }
    var builder strings.Builder
    for _, r := range base {
        if unicode.IsLetter(r) || unicode.IsDigit(r) {
            builder.WriteRune(r)
        } else {
            builder.WriteRune('_')
        }
    }
    cleaned := strings.Trim(builder.String(), "_")
    if cleaned == "" {
        return fmt.Sprintf("character_%02d", index+1)
    }
    if index >= 0 {
        return fmt.Sprintf("%s_%02d", cleaned, index+1)
    }
    return cleaned
}
