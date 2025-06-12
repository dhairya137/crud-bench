package generators

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	// Regular expressions for parsing templates
	stringRegex     = regexp.MustCompile(`string:(\d+)`)
	stringRangeRegex = regexp.MustCompile(`string:(\d+)\.\.(\d+)`)
	textRegex       = regexp.MustCompile(`text:(\d+)`)
	textRangeRegex  = regexp.MustCompile(`text:(\d+)\.\.(\d+)`)
	intRangeRegex   = regexp.MustCompile(`int:(\d+)\.\.(\d+)`)
	floatRangeRegex = regexp.MustCompile(`float:(\d+(?:\.\d+)?)\.\.(\d+(?:\.\d+)?)`)
	enumRegex       = regexp.MustCompile(`enum:(.+)`)
	intEnumRegex    = regexp.MustCompile(`int:(.+)`)
	floatEnumRegex  = regexp.MustCompile(`float:(.+)`)
)

// Initialize random seed
func init() {
	rand.Seed(time.Now().UnixNano())
}

// RandomString generates a random string of the specified length
func RandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// RandomWord generates a random word of the specified length
func RandomWord(minLen, maxLen int) string {
	length := minLen
	if maxLen > minLen {
		length = minLen + rand.Intn(maxLen-minLen+1)
	}
	return RandomString(length)
}

// RandomText generates random text made of words
func RandomText(length int) string {
	words := []string{}
	currentLength := 0
	
	for currentLength < length {
		// Generate a word between 2 and 10 characters
		wordLen := 2 + rand.Intn(9)
		if currentLength + wordLen + 1 > length {
			wordLen = length - currentLength
			if wordLen <= 0 {
				break
			}
		}
		
		word := RandomString(wordLen)
		words = append(words, word)
		currentLength += wordLen + 1 // +1 for space
	}
	
	return strings.Join(words, " ")
}

// ParseValue parses a template string and generates a value
func ParseValue(template string) interface{} {
	switch {
	case template == "int":
		return rand.Int31()
	case intRangeRegex.MatchString(template):
		matches := intRangeRegex.FindStringSubmatch(template)
		min, _ := strconv.Atoi(matches[1])
		max, _ := strconv.Atoi(matches[2])
		return min + rand.Intn(max-min+1)
	case template == "float":
		return rand.Float32()
	case floatRangeRegex.MatchString(template):
		matches := floatRangeRegex.FindStringSubmatch(template)
		min, _ := strconv.ParseFloat(matches[1], 32)
		max, _ := strconv.ParseFloat(matches[2], 32)
		return min + rand.Float64()*(max-min)
	case template == "bool":
		return rand.Intn(2) == 1
	case template == "uuid":
		return uuid.New().String()
	case template == "datetime":
		return time.Now().Format(time.RFC3339)
	case stringRegex.MatchString(template):
		matches := stringRegex.FindStringSubmatch(template)
		length, _ := strconv.Atoi(matches[1])
		return RandomString(length)
	case stringRangeRegex.MatchString(template):
		matches := stringRangeRegex.FindStringSubmatch(template)
		min, _ := strconv.Atoi(matches[1])
		max, _ := strconv.Atoi(matches[2])
		length := min + rand.Intn(max-min+1)
		return RandomString(length)
	case textRegex.MatchString(template):
		matches := textRegex.FindStringSubmatch(template)
		length, _ := strconv.Atoi(matches[1])
		return RandomText(length)
	case textRangeRegex.MatchString(template):
		matches := textRangeRegex.FindStringSubmatch(template)
		min, _ := strconv.Atoi(matches[1])
		max, _ := strconv.Atoi(matches[2])
		length := min + rand.Intn(max-min+1)
		return RandomText(length)
	case enumRegex.MatchString(template):
		matches := enumRegex.FindStringSubmatch(template)
		options := strings.Split(matches[1], ",")
		return options[rand.Intn(len(options))]
	case intEnumRegex.MatchString(template):
		matches := intEnumRegex.FindStringSubmatch(template)
		options := strings.Split(matches[1], ",")
		selected := options[rand.Intn(len(options))]
		val, _ := strconv.Atoi(selected)
		return val
	case floatEnumRegex.MatchString(template):
		matches := floatEnumRegex.FindStringSubmatch(template)
		options := strings.Split(matches[1], ",")
		selected := options[rand.Intn(len(options))]
		val, _ := strconv.ParseFloat(selected, 32)
		return val
	default:
		return template
	}
}

// ProcessTemplate processes a JSON template and replaces placeholders with random values
func ProcessTemplate(template string) (map[string]interface{}, error) {
	var data map[string]interface{}
	
	// Parse the JSON template
	if err := json.Unmarshal([]byte(template), &data); err != nil {
		return nil, fmt.Errorf("invalid JSON template: %w", err)
	}
	
	// Process the template recursively
	ProcessValue(data)
	
	return data, nil
}

// ProcessValue recursively processes values in the template
func ProcessValue(v interface{}) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		for k, v := range val {
			val[k] = ProcessValue(v)
		}
		return val
	case []interface{}:
		for i, v := range val {
			val[i] = ProcessValue(v)
		}
		return val
	case string:
		return ParseValue(val)
	default:
		return val
	}
}

// GenerateSample generates a sample value based on the template
func GenerateSample(template string) (string, error) {
	data, err := ProcessTemplate(template)
	if err != nil {
		return "", err
	}
	
	// Convert back to JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}
	
	return string(jsonData), nil
} 