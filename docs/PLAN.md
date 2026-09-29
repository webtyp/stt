---
PLAN: "feat: stt contract — Transcriber and StreamTranscriber"
TAG: v0.1.0
EXECUTOR: jules
REVIEWER: none
STATUS: review
SESSION: 2878250279206436709
PR: https://github.com/webtyp/stt/pull/1
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.
>
> **Phase 4b** of
> [`AGENT_ECOSYSTEM_MASTER_PLAN.md`](https://github.com/webtyp/agent/blob/main/docs/AGENT_ECOSYSTEM_MASTER_PLAN.md).
> **Blocked until `webtyp.com/audio` v0.1.0 is published** (phase 4a). Runs in parallel with `webtyp/tts`.

# Plan — `webtyp.com/stt`: the contract for turning speech into text

## 0. Context

**Speech-to-text (STT)**, also called automatic speech recognition (ASR), turns recorded
speech into written text. It is planned for version 2 of the webtyp agent: the person talks,
the transcript goes to `agent.Run` as if it had been typed. Version 1 is text only, but the
contract is declared now so the pipeline is wired from the start.

This plan creates only the contract. The model that implements it (candidates in
`docs/SPEECH_TO_TEXT_BROWSER_MODELS.md`) runs in the browser in Go/TinyGo, like
`webtyp/embed`, and lives in its own repository once chosen.

## Development rules (inline)

- **Contract library:** interface + value types only. No model, no browser call, no mock
  outside `_test.go` files.
- Every file compiles under `GOOS=js GOARCH=wasm` and TinyGo.
- **Never import:** `fmt`/`errors`/`strings`/`strconv` (use `webtyp.com/fmt`), `context`
  (use `webtyp.com/context`), `encoding/json`, `time`, `map[K]V`, `os`, `log`, `net/http`.
- Dependencies: `webtyp.com/audio` (v0.1.0) and `webtyp.com/context`; the test file also uses `webtyp.com/fmt`. Nothing else.
- Flat layout, tests with `testing` only. Do **not** run `gopush`/`codejob`.

## Design gate (api-design — five answers)

1. **Prior art.** **OpenAI Whisper API**: `transcriptions.create(file)` → `{text, language}`.
   **Google Cloud Speech-to-Text**: `Recognize(config, audio)` → alternatives with transcript.
   **Web Speech API**: `SpeechRecognition` (event-based, browser-owned model).
   **whisper.cpp**: `whisper_full(ctx, params, samples)` on float32 16 kHz PCM. We take
   Whisper's vocabulary (*transcribe*, *text + language*) and whisper.cpp's input (float32
   PCM), which is what `audio.PCM` already is. For streaming (partial text while the
   person is still talking), **Vosk** (`AcceptWaveform` → `PartialResult` → `FinalResult`) and
   **sherpa-onnx** `OnlineStream` (`AcceptWaveform`, `InputFinished`) both use the same shape:
   open a stream, push audio pieces, finish. We use it as `NewStream` → `Write` → `Close`, the
   verbs of Go's `io.WriteCloser`, which any Go developer already knows. It is a separate
   interface (`StreamTranscriber`), not a mode flag, so a batch-only model is never forced to
   fake it. Nothing consumes it in version 1. It is declared now so the voice loop is wired
   from the start.
2. **Novice-name test.** `stt.Transcriber.Transcribe(ctx, speech)` returning `stt.Transcript{Text, Language}`
   reads as "transcribe this speech". No abbreviations beyond the repository name.
3. **Complexity ledger.** Concepts +4 (`Transcriber`, `Transcript`, `StreamTranscriber`, `Stream`) / −0. Ways to do the same
   thing +0: nothing else transcribes today.
4. **Where it belongs.** STT is its own concern with its own models and licences. It is
   separate from text-to-speech (`webtyp/tts`), because an app can dictate without ever
   speaking back. The contract repo holds no implementation, so importing it adds no model to
   a binary.
5. **What it deletes.** Nothing. This is new capability.

## Stage 1 — the contract

**`stt.go`**

```go
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
```

## Stage 2 — consumer-shaped test

**File:** `example_test.go`, package `stt_test`. Declare in the file a
`type fixedTranscriber struct{ text string }` whose `Transcribe` rejects any `speech` with
`SampleRate != 16000` using `fmt.Err("stt: expected 16000 Hz")`, and otherwise returns
`Transcript{Text: t.text, Language: "es"}`.

- `Example_dictation`: one second of silence at 16 kHz mono → prints the transcript text.
  `// Output: reserva una hora`
- `TestTranscribe_RejectsWrongRate`: 48000 Hz → error.
- `TestStream_PartialThenFinal`: a `wordStream` in the test file whose `Write` appends one
  word per call from a fixed list and returns the words so far, and whose `Close` returns
  them all. Three writes return `"reserva"`, `"reserva una"`, `"reserva una hora"`; `Close`
  returns `"reserva una hora"`; a `Write` after `Close` returns an error. Also assert that
  `var _ stt.StreamTranscriber = wordTranscriber{}` compiles.

## Stages

| Stage | Files | Acceptance |
|---|---|---|
| 1 | `stt.go`, `go.mod` | builds for host, wasm, TinyGo |
| 2 | `example_test.go` | passes under `gotest` and `gotest -tinygo` |
