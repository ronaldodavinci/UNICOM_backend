package utils

import (
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
	"unicode"
)

// ProfanityFilter provides simple profanity masking using a list of patterns.
// It replaces each matched word with a string of '*' having the same rune length.
// Notes:
//   - For ASCII words, boundaries are applied to avoid partial matches inside other words.
//   - For non-ASCII scripts (e.g., Thai), exact substring matching is used (no boundaries),
//     which is more practical for scripts without spaces between words.
type ProfanityFilter struct {
	patterns []*regexp.Regexp
}

var (
	defaultFilter     *ProfanityFilter
	defaultFilterOnce sync.Once
)

// DefaultBannedWords is a small starter list. Extend via env or config if needed.
var DefaultBannedWords = []string{
	// ===== English (case-insensitive) =====
	"fuck", "fucking", "fucker", "motherfucker", "shit", "bullshit",
	"bastard", "bitch", "sonofabitch", "dick", "cock", "pussy", "cunt",
	"asshole", "dumbass", "jackass", "retard", "moron", "slut", "whore",
	"nigger", "faggot", "jerk", "idiot", "crap", "douche", "douchebag",
	"wanker", "twat", "prick", "arsehole", "bloody", "bugger", "bollocks",
	"damn", "dammit", "piss", "pissed", "hell", "screw", "screwed",
	"cocksucker", "ballsack", "nutsack", "buttfuck", "butthole", "shithead",
	"shitface", "dipshit", "dumbfuck", "numbnuts", "cocks", "cum", "cumshot",
	"milf", "gayass", "dildo", "porn", "sex", "sexual", "orgasm", "anal",
	"penetrate", "rapist", "rape", "suck", "sucker", "deepthroat",
	"jerkoff", "masturbate", "masturbation", "wank", "handjob", "blowjob",
	"hardcore", "fuckface", "shitbag", "cockhead", "nuts", "fart", "pisshead",
	"arse", "arseface", "bollock", "buggered", "shag", "twit", "tosser",

	// ===== Thai (case-sensitive, ไทยไม่มีพิมพ์เล็ก/ใหญ่) =====
	"หี", "ควย", "เย็ด", "เหี้ย", "สัส", "แม่ง", "พ่อง", "พ่อมึง",
	"แม่มึง", "มึง", "เงี่ยน", "สัด", "เวร", "ระยำ",
	"ชาติหมา", "ตอแหล", "สถุน", "เลว", "กะหรี่", "ควาย",
	"อัปรีย์", "ห่า", "ฟาย", "ไอสัส", "ไอเหี้ย", "ไอเวร",
	"ไอโง่", "ไอบ้า", "ไอควาย", "ไอเลว", "ไอพ่อง", "เหี้ยแม่ง",
	"แม่งเอ๊ย", "สันดาน", "โง่", "ชิบหาย", "อีห่า", "อีสัด", "อีเหี้ย",
	"อีบ้า", "อีควาย", "อีเลว", "อีดอก", "อีเปรต", "อีเชี้ย", "ดอกทอง",
	"เอ๋อ", "กะโปก", "จู๋", "จิ๋ม", "กู", "มึง", "แตด",
	"ชาติชั่ว", "ไอขยะ", "ขยะสังคม", "ขี้ข้า", "เฮงซวย",
	"กะโหลกกะลา", "ส้นตีน", "หัวควย", "ปัญญาอ่อน", "กาก",
	"สมองกลวง", "ไอขี้แพ้", "ส้นตีน", "ควยแตก", "น้ำแตก",
	"หำ", "ค.ว.ย", "ห.ย.", "ค.ว.ย.", "เหรี้ย", "เ-หี้ย",
	"เ_หี้ย", "ห_ี", "ห-ี", "ค-วย", "ค_วย", "ค*ย", "ค ว ย",
}

// MaskProfanity masks profanity in the given text using the default filter.
// It reads additional words from env PROFANITY_WORDS (comma-separated) if present.
func MaskProfanity(s string) string {
	if len(s) == 0 {
		return s
	}
	defaultFilterOnce.Do(func() {
		words := make([]string, 0, len(DefaultBannedWords))
		words = append(words, DefaultBannedWords...)
		if extra := strings.TrimSpace(os.Getenv("PROFANITY_WORDS")); extra != "" {
			for _, w := range strings.Split(extra, ",") {
				w = strings.TrimSpace(w)
				if w != "" {
					words = append(words, w)
				}
			}
		}
		// Ensure key Thai words present to avoid partial masking (e.g., หี inside เหี้ย)
		words = append(words, "เหี้ย", "หี")
		defaultFilter = NewProfanityFilter(words)
	})
	return defaultFilter.Mask(s)
}

// NewProfanityFilter builds a filter from a list of banned words.
// Words containing only ASCII letters/digits get word boundaries.
// Others (e.g., Thai) are matched as-is (case-sensitive).
func NewProfanityFilter(words []string) *ProfanityFilter {
	// Trim + de-dup
	uniq := make([]string, 0, len(words))
	seen := map[string]struct{}{}
	for _, w := range words {
		w = strings.TrimSpace(w)
		if w == "" {
			continue
		}
		if _, ok := seen[w]; ok {
			continue
		}
		seen[w] = struct{}{}
		uniq = append(uniq, w)
	}
	// Sort by rune length desc so longer Thai words match before shorter substrings
	sort.Slice(uniq, func(i, j int) bool {
		return len([]rune(uniq[i])) > len([]rune(uniq[j]))
	})
	var pats []*regexp.Regexp
	for _, w := range uniq {
		onlyASCIIWord := isASCIIWord(w)
		var pattern string
		if onlyASCIIWord {
			// (?i) for case-insensitive; \b boundaries around the word
			pattern = `(?i)\b` + regexp.QuoteMeta(w) + `\b`
		} else {
			// Non-ASCII word: match as substring (no boundaries)
			pattern = regexp.QuoteMeta(w)
		}
		re := regexp.MustCompile(pattern)
		pats = append(pats, re)
	}
	return &ProfanityFilter{patterns: pats}
}

// Mask walks through patterns and masks matches with '*'.
func (pf *ProfanityFilter) Mask(s string) string {
	if pf == nil || len(pf.patterns) == 0 || s == "" {
		return s
	}
	out := s
	for _, re := range pf.patterns {
		out = re.ReplaceAllStringFunc(out, func(m string) string {
			// Count runes to avoid breaking multibyte characters.
			n := len([]rune(m))
			if n <= 0 {
				return m
			}
			return strings.Repeat("*", n)
		})
	}
	return out
}

func isASCIIWord(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r > unicode.MaxASCII {
			return false
		}
		if !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-') {
			return false
		}
	}
	return true
}
