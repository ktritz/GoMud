package parser

import (
	"fmt"
	"strconv"
	"strings"
)

// Filler words that are stripped from structured input when they appear
// before the first noun (e.g., "look at the goblin" -> "look goblin").
// These are NOT stripped when they appear as structural prepositions
// between two noun phrases (e.g., "unlock door with key" preserves "with").
var fillerWords = map[string]bool{
	"the": true,
	"a":   true,
	"an":  true,
	"at":  true,
	"my":  true,
	"its": true,
	"ye":  true,
}

// Structural prepositions that can separate target from instrument.
// These are preserved as parsed.Preposition when they appear between two noun phrases.
var structuralPrepositions = map[string]bool{
	"with":    true,
	"from":    true,
	"in":      true,
	"into":    true,
	"on":      true,
	"onto":    true,
	"to":      true,
	"about":   true,
	"under":   true,
	"upon":    true,
	"toward":  true,
	"towards": true,
	"over":    true,
	"using":   true,
	"for":     true,
}

// Implicit prepositions — when these verbs have 2+ tokens but no explicit
// preposition, the last token is treated as the instrument (recipient/source).
// E.g., "give sword merchant" implies "give sword to merchant".
// For multi-word recipient names, use the explicit preposition: "give sword to old merchant"
var implicitPrepositions = map[string]string{
	"give": "to",
	"show": "to",
	"buy":  "from",
}

// Multi-word command collapses — checked before verb/rest split.
var multiWordCommands = map[string]string{
	"pick up":  "get",
	"put on":   "equip",
	"take off": "remove",
	"look at":  "look",
	"look in":  "look",
}

// Token synonyms — applied to individual tokens in the rest string.
// Verb synonyms are handled by the existing keyword alias system.
var tokenSynonyms = map[string]string{
	// These supplement the existing command-aliases in keywords.yaml
	// which handle verb-level synonyms. These handle noun-level synonyms.
}

// Command classification — determines how the parser treats the rest string.
var commandClasses = map[string]CommandClass{
	// Freeform — don't parse the rest, it's player text
	"say":       ClassFreeform,
	"shout":     ClassFreeform,
	"yell":      ClassFreeform,
	"emote":     ClassFreeform,
	"broadcast": ClassFreeform,
	"whisper":   ClassFreeform,
	"print":     ClassFreeform,
	"motd":      ClassFreeform,
	// Passthrough — admin/system commands
	"server":    ClassPassthrough,
	"build":     ClassPassthrough,
	"room":      ClassPassthrough,
	"spawn":     ClassPassthrough,
	"modify":    ClassPassthrough,
	"questtoken": ClassPassthrough,
	"zap":       ClassPassthrough,
}

// GetCommandClass returns the parsing class for a command verb.
func GetCommandClass(verb string) CommandClass {
	if class, ok := commandClasses[strings.ToLower(verb)]; ok {
		return class
	}
	return ClassStructured
}

// TryMultiWordCollapse checks if the first two words of input form a
// multi-word command and collapses them into a single verb.
// Returns (verb, rest, collapsed).
func TryMultiWordCollapse(input string) (string, string, bool) {
	lower := strings.ToLower(input)
	for multi, replacement := range multiWordCommands {
		if strings.HasPrefix(lower, multi+" ") {
			rest := strings.TrimSpace(input[len(multi):])
			return replacement, rest, true
		}
		if lower == multi {
			return replacement, "", true
		}
	}
	return "", "", false
}

// TryPrefixMatch attempts to match an unrecognized command by prefix
// against a list of known commands. Returns the match if exactly one
// command matches the prefix, otherwise returns empty string.
func TryPrefixMatch(input string, knownCommands []string) string {
	input = strings.ToLower(input)
	if len(input) < 2 {
		return "" // Don't match single characters
	}

	var matches []string
	for _, cmd := range knownCommands {
		if strings.HasPrefix(cmd, input) {
			matches = append(matches, cmd)
		}
	}

	if len(matches) == 1 {
		return matches[0]
	}
	return ""
}

// tokenize splits input into tokens, respecting quoted strings.
func tokenize(input string) []string {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil
	}

	var tokens []string
	var current strings.Builder
	inQuote := false

	for _, ch := range input {
		switch {
		case ch == '"' || ch == '\'':
			inQuote = !inQuote
			current.WriteRune(ch)
		case ch == ' ' && !inQuote:
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(ch)
		}
	}
	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}

	return tokens
}

// stripFillers removes filler words from a token list.
func stripFillers(tokens []string) []string {
	result := make([]string, 0, len(tokens))
	for _, t := range tokens {
		if !fillerWords[strings.ToLower(t)] {
			result = append(result, t)
		}
	}
	return result
}

// isStructuralPreposition returns true if the token is a preposition
// that can separate target from instrument noun phrases.
func isStructuralPreposition(token string) bool {
	return structuralPrepositions[strings.ToLower(token)]
}

// resolveTokenSynonym resolves a token to its canonical form if a synonym exists.
func resolveTokenSynonym(token string) string {
	if syn, ok := tokenSynonyms[strings.ToLower(token)]; ok {
		return syn
	}
	return token
}

// parseNounPhrase builds a NounPhrase from a list of tokens.
// The last token is the noun, everything before it is adjectives.
// Handles #N disambiguation, @/#/* prefixes, quantities, and "all".
func parseNounPhrase(tokens []string) NounPhrase {
	if len(tokens) == 0 {
		return NounPhrase{}
	}

	raw := strings.Join(tokens, " ")

	// Handle "all" keyword alone
	if len(tokens) == 1 && strings.ToLower(tokens[0]) == "all" {
		return NounPhrase{Raw: raw, Noun: "all", All: true}
	}

	// Handle special prefixes (@userId, #mobInstanceId, *random) — pass through
	if len(tokens) == 1 && len(tokens[0]) > 1 {
		first := tokens[0][0]
		if first == '@' || first == '#' || first == '*' {
			return NounPhrase{Raw: raw, Noun: tokens[0]}
		}
	}

	// Check if first token is a number (quantity like "5 gold")
	quantity := 0
	startIdx := 0
	if len(tokens) > 1 {
		if n, err := strconv.Atoi(tokens[0]); err == nil && n > 0 {
			quantity = n
			startIdx = 1
		}
	}

	// Check if first token (after quantity) is "all"
	isAll := false
	if startIdx < len(tokens) && strings.ToLower(tokens[startIdx]) == "all" {
		isAll = true
		startIdx++
	}

	remaining := tokens[startIdx:]
	if len(remaining) == 0 {
		return NounPhrase{Raw: raw, All: isAll, Quantity: quantity}
	}

	// Last remaining token is the noun, everything else is adjectives
	noun := remaining[len(remaining)-1]
	adjectives := make([]string, 0)
	if len(remaining) > 1 {
		adjectives = append(adjectives, remaining[:len(remaining)-1]...)
	}

	// Check for #N disambiguation on the noun
	index := 0
	if hashIdx := strings.LastIndex(noun, "#"); hashIdx > 0 {
		numStr := noun[hashIdx+1:]
		if n, err := strconv.Atoi(numStr); err == nil && n > 0 {
			index = n
			noun = noun[:hashIdx]
		}
	}

	return NounPhrase{
		Raw:        raw,
		Adjectives: adjectives,
		Noun:       noun,
		Index:      index,
		Quantity:   quantity,
		All:        isAll,
	}
}

// FormatNounPhrase reconstructs a noun phrase string, including #N if present.
func FormatNounPhrase(np NounPhrase) string {
	if np.IsEmpty() {
		return ""
	}
	result := ""
	if len(np.Adjectives) > 0 {
		result = strings.Join(np.Adjectives, " ") + " "
	}
	result += np.Noun
	if np.Index > 0 {
		result += fmt.Sprintf("#%d", np.Index)
	}
	return result
}
