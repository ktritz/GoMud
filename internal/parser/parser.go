package parser

import (
	"strings"

	"github.com/GoMudEngine/GoMud/internal/users"
)

const tempDataKey = `parsedInput`

// CommandClass determines how the parser treats the rest string.
type CommandClass int

const (
	ClassStructured  CommandClass = iota // Full parsing: tokenize, strip fillers, resolve nouns
	ClassFreeform                        // No parsing: say, emote, shout, broadcast, whisper
	ClassPassthrough                     // Minimal parsing: admin commands, macros
)

// ParsedInput is the structured result of parsing player input.
type ParsedInput struct {
	RawInput    string     // The original unmodified input
	Verb        string     // Resolved command verb
	Rest        string     // Cleaned rest string (fillers stripped, synonyms resolved)
	OrigRest    string     // The original rest string before any modification
	Target      NounPhrase // Primary target (e.g., "red sword", "goblin")
	Instrument  NounPhrase // Secondary object after preposition (e.g., "key" in "unlock door with key")
	Preposition string     // The preposition connecting target and instrument
	Tokens      []string   // All tokens after filler removal
	Extra       string     // Remaining unparsed text (for freeform commands)
}

// NounPhrase represents a noun with optional adjective(s) and disambiguation index.
type NounPhrase struct {
	Raw        string   // The original text of this noun phrase
	Adjectives []string // Leading adjectives
	Noun       string   // The head noun
	Index      int      // Disambiguation number from #N syntax (default 0 = no disambiguation)
	Quantity   int      // Numeric quantity (e.g., "5 gold" -> Quantity=5, Noun="gold"), 0 = not specified
	All        bool     // True if "all" keyword was used (e.g., "get all")
}

func (np NounPhrase) IsEmpty() bool {
	return np.Noun == ""
}

// String reconstructs the noun phrase for passing to existing FindByName/FindMatchIn.
func (np NounPhrase) String() string {
	if np.IsEmpty() {
		return np.Raw
	}
	result := strings.Join(append(np.Adjectives, np.Noun), " ")
	if np.Index > 0 {
		result += "#" + strings.TrimLeft(string(rune('0'+np.Index)), "")
	}
	return result
}

// Parse takes a resolved verb and the rest string, and produces a ParsedInput.
func Parse(verb string, rest string, class CommandClass) *ParsedInput {
	parsed := &ParsedInput{
		RawInput: verb + " " + rest,
		Verb:     verb,
		OrigRest: rest,
	}

	switch class {
	case ClassFreeform:
		// Don't touch the text — it's chat/emote content
		parsed.Rest = rest
		parsed.Extra = rest
		return parsed

	case ClassPassthrough:
		// Minimal processing — just tokenize, don't strip fillers
		parsed.Rest = rest
		parsed.Tokens = tokenize(rest)
		if len(parsed.Tokens) > 0 {
			parsed.Target = parseNounPhrase(parsed.Tokens)
		}
		return parsed

	case ClassStructured:
		// Full parsing
		return parseStructured(parsed, rest)
	}

	parsed.Rest = rest
	return parsed
}

func parseStructured(parsed *ParsedInput, rest string) *ParsedInput {
	if rest == "" {
		parsed.Rest = ""
		return parsed
	}

	tokens := tokenize(rest)
	if len(tokens) == 0 {
		parsed.Rest = ""
		return parsed
	}

	// Resolve synonyms on each token
	for i, t := range tokens {
		tokens[i] = resolveTokenSynonym(t)
	}

	// Find preposition that splits target from instrument
	prepIdx := -1
	for i, t := range tokens {
		if i == 0 {
			continue // First token is always part of the target
		}
		if isStructuralPreposition(t) {
			prepIdx = i
			break
		}
	}

	if prepIdx > 0 {
		// Split into target tokens and instrument tokens
		targetTokens := stripFillers(tokens[:prepIdx])
		parsed.Preposition = tokens[prepIdx]
		instrumentTokens := stripFillers(tokens[prepIdx+1:])

		parsed.Target = parseNounPhrase(targetTokens)
		parsed.Instrument = parseNounPhrase(instrumentTokens)
		parsed.Tokens = append(targetTokens, instrumentTokens...)
	} else {
		// No preposition found
		cleanTokens := stripFillers(tokens)

		// Check for implicit preposition (e.g., "give sword merchant" implies "to")
		if implicitPrep, ok := implicitPrepositions[parsed.Verb]; ok && len(cleanTokens) >= 2 {
			// Last token is the instrument, everything else is the target
			targetTokens := cleanTokens[:len(cleanTokens)-1]
			instrumentTokens := cleanTokens[len(cleanTokens)-1:]

			parsed.Preposition = implicitPrep
			parsed.Target = parseNounPhrase(targetTokens)
			parsed.Instrument = parseNounPhrase(instrumentTokens)
			parsed.Tokens = cleanTokens
		} else {
			parsed.Target = parseNounPhrase(cleanTokens)
			parsed.Tokens = cleanTokens
		}
	}

	// Rebuild rest string from cleaned tokens
	parsed.Rest = rebuildRest(parsed)

	return parsed
}

func rebuildRest(parsed *ParsedInput) string {
	parts := []string{}

	if !parsed.Target.IsEmpty() {
		parts = append(parts, parsed.Target.String())
	}

	if parsed.Preposition != "" {
		parts = append(parts, parsed.Preposition)
	}

	if !parsed.Instrument.IsEmpty() {
		parts = append(parts, parsed.Instrument.String())
	}

	if len(parts) == 0 && len(parsed.Tokens) > 0 {
		return strings.Join(parsed.Tokens, " ")
	}

	return strings.Join(parts, " ")
}

// GetParsedInput retrieves the parsed input from the user's temp data store.
func GetParsedInput(user *users.UserRecord) *ParsedInput {
	if data := user.GetTempData(tempDataKey); data != nil {
		if pi, ok := data.(*ParsedInput); ok {
			return pi
		}
	}
	return nil
}

// StoreParsedInput stores the parsed input on the user for handler access.
func StoreParsedInput(user *users.UserRecord, parsed *ParsedInput) {
	user.SetTempData(tempDataKey, parsed)
}
