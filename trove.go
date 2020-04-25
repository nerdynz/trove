package trove

import (
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	dotenv "github.com/joho/godotenv"
)

// Settings - common settings used around the site. Currently loaded into the datastore object
type Settings struct {
	// ServerIsLVE bool
	// ServerIsDEV bool
	// ServerIs             string
	// DSN                  string
	// CanonicalURL         string
	// WebsiteBaseURL       string
	// ImageBaseURL         string
	// Sitename             string
	// EncKey               string
	// ServerPort string
	// AttachmentsFolder    string
	// MaxImageWidth        int
	// IsSecured            bool
	// Proto                string
	// SlackLogURL          string
	// CheckCSRFViaReferrer bool
	// EmailFromName        string
	// EmailFromEmail       string
	// IsSiteBound          bool
	// CacheNamespace       string
	// LoggingEnabled       bool
	bools   map[string]bool
	bLock   sync.RWMutex
	strings map[string]string
	sLock   sync.RWMutex
}

func Load() *Settings {
	err := dotenv.Load()
	if err != nil {
		if strings.Contains(err.Error(), "no such file or directory") {
			// assume no dotenv file is present... i.e live
		} else {
			panic(err)
		}
	}
	s := &Settings{}
	s.bLock = sync.RWMutex{}
	s.sLock = sync.RWMutex{}

	s.bools = map[string]bool{}
	s.strings = map[string]string{}
	// s.ServerIsDEV = (os.Getenv("IS_DEV") == "true")
	// s.ServerIsLVE = !s.ServerIsDEV
	if os.Getenv("IS_DEV") == "true" {
		s.strings["SERVER_IS"] = "DEV"
		s.bools["SERVER_IS_DEV"] = true
	} else {
		s.strings["SERVER_IS"] = "LVE"
		s.bools["SERVER_IS_LVE"] = true
	}
	s.strings["DSN"] = os.Getenv("DATABASE_URL")

	imgWidth := os.Getenv("MAX_IMAGE_WIDTH")
	if imgWidth == "" {
		s.strings["MAX_IMAGE_WIDTH"] = imgWidth
	}

	canonicalURL := strings.ToLower(os.Getenv("CANONICAL_URL"))
	if canonicalURL != "" {
		s.strings["CANONICAL_URL"] = canonicalURL
	}

	s.bools["IS_SECURED"] = (strings.ToLower(os.Getenv("IS_HTTPS")) == "true")
	s.strings["PROTO"] = "http://"
	if s.IsProduction() {
		s.strings["PROTO"] = "https://"
	}
	websiteBaseURL := os.Getenv("WEBSITE_BASE_URL")
	if websiteBaseURL == "" {
		s.strings["WEBSITE_BASE_URL"] = s.strings["PROTO"] + s.strings["CANONICAL_URL"] + "/"
	}

	if s.Get("REDIS_URL") != "" {
		s.strings["CACHE_URL"] = s.Get("REDIS_URL")
	}

	return s
}

// GetDuration gets a duration from either a duration formatted string (time.ParseDuration) or a string ending in day(s) e.g. 30days
func (s *Settings) GetDuration(setting string) time.Duration {
	val := s.Get(setting)

	oddMod := ""
	if strings.Contains(val, "day") {
		oddMod = "day"
	}
	// if strings.Contains(val, "month") { // hmmmm
	// 	oddMod = "month"
	// }
	// if strings.Contains(val, "year") { // hmmmm
	// 	oddMod = "year"
	// }

	if oddMod != "" {
		val = strings.ToLower(val)
		re := regexp.MustCompile("[0-9]+")
		durStr := re.FindAllString(val, 1)
		durNum, err := strconv.Atoi(durStr[0])
		if err != nil {
			return -1
		}
		if oddMod == "day" {
			return time.Hour * time.Duration(durNum) * 24
		}
		return -1
	} else {
		dur, err := time.ParseDuration(val)
		if err != nil {
			return -1
		}
		return dur
	}
	return -1
}

func (s *Settings) GetBool(setting string) bool {
	val, ok := s.bools[setting]
	if !ok {
		s.bLock.Lock() // map is shared across the whole application, so it needs to be
		defer s.bLock.Unlock()

		newVal := strings.ToLower(os.Getenv(setting)) == "true" || strings.ToLower(os.Getenv(setting)) == "1"
		s.bools[setting] = newVal
		val = newVal
	}
	return val
}

// Get gets and caches an ENV setting. theDefault will be used if setting is blank and is optional
func (s *Settings) Get(setting string) string {
	return s.GetWithDefault(setting, "")
}

func (s *Settings) GetWithDefault(setting string, theDefault string) string {
	val, ok := s.strings[setting]
	if !ok {
		s.sLock.Lock() // map is shared across the whole application, so it needs to be
		defer s.sLock.Unlock()

		newVal := os.Getenv(setting)
		if newVal == "" && theDefault != "" {
			return theDefault
		}
		s.strings[setting] = newVal
		val = newVal
	}
	return val
}

func (s *Settings) IsProduction() bool {
	return s.GetBool("SERVER_IS_LVE")
}
func (s *Settings) IsDevelopment() bool {
	return s.GetBool("SERVER_IS_DEV")
}
