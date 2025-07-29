package transcription

import (
	"regexp"
	"strings"
)

// TextFormatter handles formatting transcribed text into readable paragraphs
type TextFormatter struct {
	targetWordCount                int
	maxSentencesPerChunk           int
	minWordsForSignificantSentence int
}

// NewTextFormatter creates a new text formatter with default settings
func NewTextFormatter() *TextFormatter {
	return &TextFormatter{
		targetWordCount:                50, // Target ~50 words per paragraph
		maxSentencesPerChunk:           4,  // Maximum 4 sentences per paragraph
		minWordsForSignificantSentence: 4,  // Sentences with 4+ words are "significant"
	}
}

// Format takes raw transcription text and formats it into readable paragraphs
func (f *TextFormatter) Format(text string) string {
	if strings.TrimSpace(text) == "" {
		return ""
	}

	// Split text into sentences using punctuation
	sentences := f.splitIntoSentences(text)
	if len(sentences) == 0 {
		return text
	}

	var finalFormattedText strings.Builder

	processedSentenceIndex := 0

	for processedSentenceIndex < len(sentences) {
		var currentChunkSentences []string

		currentChunkWordCount := 0
		currentChunkSignificantSentenceCount := 0

		// Build a tentative chunk based on target word count
		for i := processedSentenceIndex; i < len(sentences); i++ {
			sentence := sentences[i]
			wordsInSentence := f.countWords(sentence)

			currentChunkSentences = append(currentChunkSentences, sentence)
			currentChunkWordCount += wordsInSentence

			if wordsInSentence >= f.minWordsForSignificantSentence {
				currentChunkSignificantSentenceCount++
			}

			// Stop if we've reached our target word count
			if currentChunkWordCount >= f.targetWordCount {
				break
			}
		}

		// Apply max sentences per chunk rule based on significant sentences
		var sentencesForFinalChunk []string

		if currentChunkSignificantSentenceCount > f.maxSentencesPerChunk {
			significantSentenceCount := 0

			for _, sentence := range currentChunkSentences {
				sentencesForFinalChunk = append(sentencesForFinalChunk, sentence)

				wordsInSentence := f.countWords(sentence)
				if wordsInSentence >= f.minWordsForSignificantSentence {
					significantSentenceCount++
					if significantSentenceCount >= f.maxSentencesPerChunk {
						break
					}
				}
			}
		} else {
			sentencesForFinalChunk = currentChunkSentences
		}

		// Add the chunk to final text
		if len(sentencesForFinalChunk) > 0 {
			chunkText := strings.Join(sentencesForFinalChunk, " ")
			chunkText = f.cleanText(chunkText)

			if finalFormattedText.Len() > 0 {
				finalFormattedText.WriteString("\n\n")
			}

			finalFormattedText.WriteString(chunkText)

			processedSentenceIndex += len(sentencesForFinalChunk)
		} else {
			// Safety break to avoid infinite loop
			break
		}
	}

	return strings.TrimSpace(finalFormattedText.String())
}

// splitIntoSentences splits text into sentences using punctuation patterns
func (f *TextFormatter) splitIntoSentences(text string) []string {
	// Clean up the text first
	text = strings.TrimSpace(text)
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")

	// Split on sentence-ending punctuation followed by whitespace and capital letter
	// This regex looks for: . ! ? followed by space(s) and capital letter
	sentenceRegex := regexp.MustCompile(`([.!?]+)\s+([A-Z])`)

	// Replace matches with sentence ending + newline + capital letter
	text = sentenceRegex.ReplaceAllString(text, "$1\n$2")

	// Split on newlines and clean up
	rawSentences := strings.Split(text, "\n")

	var sentences []string

	for _, sentence := range rawSentences {
		sentence = strings.TrimSpace(sentence)
		if sentence != "" {
			sentences = append(sentences, sentence)
		}
	}

	// If no sentence splits were found, treat the whole text as one sentence
	if len(sentences) <= 1 && len(rawSentences) == 1 {
		sentences = []string{text}
	}

	return sentences
}

// countWords counts the number of words in a sentence
func (f *TextFormatter) countWords(sentence string) int {
	sentence = strings.TrimSpace(sentence)
	if sentence == "" {
		return 0
	}

	// Split on whitespace and count non-empty parts
	words := strings.Fields(sentence)

	return len(words)
}

// cleanText performs basic text cleanup
func (f *TextFormatter) cleanText(text string) string {
	// Remove extra whitespace
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")

	// Clean up common transcription artifacts
	text = strings.ReplaceAll(text, " ,", ",")
	text = strings.ReplaceAll(text, " .", ".")
	text = strings.ReplaceAll(text, " !", "!")
	text = strings.ReplaceAll(text, " ?", "?")

	// Remove multiple consecutive punctuation marks
	text = regexp.MustCompile(`[.]{2,}`).ReplaceAllString(text, ".")
	text = regexp.MustCompile(`[!]{2,}`).ReplaceAllString(text, "!")
	text = regexp.MustCompile(`[?]{2,}`).ReplaceAllString(text, "?")

	return strings.TrimSpace(text)
}
