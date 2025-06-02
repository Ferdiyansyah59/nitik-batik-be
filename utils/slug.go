package utils

import (
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/gosimple/slug"
)

// GenerateSlug creates a slug from the given string
func GenerateSlug(title string, entityType string) string {
	// Create a slug from the title
	slugStr := slug.Make(title)
	
	// Ensure slug is not empty
	if slugStr == "" {
		// If the title only contains non-alphanumeric characters,
		// generate a random string instead
		return fmt.Sprintf("%s-%s", entityType, GenerateRandomString(8))
	}
	
	return slugStr
}

// EnsureUniqueSlug makes sure the slug is unique by appending a number if needed
func EnsureUniqueSlug(baseSlug string, checkExists func(string) bool) string {
	slugStr := baseSlug
	counter := 1
	
	// Keep checking until we find a unique slug
	for checkExists(slugStr) {
		// If the slug already exists, append a counter
		slugStr = fmt.Sprintf("%s-%d", baseSlug, counter)
		counter++
		
		// Safety check to avoid infinite loops
		if counter > 100 {
			// If we can't find a unique slug after 100 attempts, use a timestamp
			return fmt.Sprintf("%s-%d", baseSlug, time.Now().UnixNano())
		}
	}
	
	return slugStr
}

// GenerateRandomString generates a random string of the specified length
func GenerateRandomString(length int) string {
	// For simplicity, let's use a timestamp-based approach
	timestamp := strings.Replace(time.Now().Format("20060102150405.000000"), ".", "", -1)
	
	// Use the last 'length' characters
	if len(timestamp) > length {
		return timestamp[len(timestamp)-length:]
	}
	
	// If timestamp is shorter than the requested length, pad with zeros
	return timestamp + strings.Repeat("0", length-len(timestamp))
}

// SanitizeSlug removes unwanted characters from a slug
func SanitizeSlug(input string) string {
	// Convert to lowercase
	input = strings.ToLower(input)
	
	// Replace spaces with hyphens
	input = strings.ReplaceAll(input, " ", "-")
	
	// Remove all non-alphanumeric characters except hyphens
	reg := regexp.MustCompile("[^a-z0-9-]")
	input = reg.ReplaceAllString(input, "")
	
	// Replace multiple hyphens with a single one
	reg = regexp.MustCompile("-+")
	input = reg.ReplaceAllString(input, "-")
	
	// Trim hyphens from beginning and end
	input = strings.Trim(input, "-")
	
	// If the slug is empty, return a default value
	if input == "" {
		return "article"
	}
	
	return input
}

// GenerateExcerpt creates an excerpt from the description
func GenerateExcerpt(description string, maxLength int) string {
	if len(description) <= maxLength {
		return description
	}
	
	// Find the last space before maxLength
	lastSpace := strings.LastIndex(description[:maxLength], " ")
	if lastSpace < 0 {
		lastSpace = maxLength
	}
	
	// Extract the excerpt
	excerpt := description[:lastSpace]
	
	// Add ellipsis if the excerpt is shorter than the description
	if len(excerpt) < len(description) {
		excerpt += "..."
	}
	
	return excerpt
}

// TruncateString truncates a string to the specified length without cutting words
func TruncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	
	// Find the last space before maxLength
	lastSpace := strings.LastIndex(s[:maxLength], " ")
	if lastSpace < 0 {
		// If no space is found, just cut at maxLength
		return s[:maxLength] + "..."
	}
	
	// Cut at the last space and add ellipsis
	return s[:lastSpace] + "..."
}

// IsAlphanumeric checks if a string contains only alphanumeric characters
func IsAlphanumeric(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) {
			return false
		}
	}
	return true
}