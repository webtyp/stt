// Package stt is the speech-to-text contract: PCM audio in, text out. Implementations live
// in their own repositories.
package stt

import (
	"webtyp.com/audio"
	"webtyp.com/context"
)

// Transcriber turns recorded speech into text.
type Transcriber interface {
	// Transcribe returns the words spoken in speech. Implementations document the sample rate
	// they accept (speech models usually expect 16000 Hz mono) and return an error for any other.
	Transcribe(ctx *context.Context, speech audio.PCM) (Transcript, error)
}

// StreamTranscriber transcribes speech while it is still being recorded.
type StreamTranscriber interface {
	NewStream(ctx *context.Context) (Stream, error)
}

// Stream is one utterance being transcribed as it arrives.
type Stream interface {
	// Write adds the next piece of speech and returns the transcript so far. The partial text
	// may still change with later pieces.
	Write(speech audio.PCM) (Transcript, error)

	// Close ends the utterance and returns the final transcript. Write after Close is an error.
	Close() (Transcript, error)
}

// Transcript is the text recognised in one piece of speech.
type Transcript struct {
	Text     string
	Language string // BCP 47 tag of the spoken language, e.g. "es"; empty when unknown
}
