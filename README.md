# stt
<img src="docs/img/badges.svg">

The speech-to-text contract of webtyp: `Transcriber.Transcribe(ctx, audio.PCM)` → `Transcript{Text, Language}`. It is a contract only. The model that implements it runs in
the browser in Go/TinyGo and lives in its own repository.

## Getting started

| I want to… | Use |
|---|---|
| turn the microphone into text | `stt.Transcriber` |
| show text while the person is still talking | `stt.StreamTranscriber` → `Stream.Write` / `Close` (version 2) |
| implement a model for it | implement the interface in your own repository |
| pass audio around | `webtyp.com/audio` (`audio.PCM`) |

## Documentation

- [Architecture](docs/ARCHITECTURE.md): the contract, where it sits in the voice pipeline, and the candidate models.
- [Browser model research](docs/SPEECH_TO_TEXT_BROWSER_MODELS.md): the model comparison this project started from.
- [Agent guide](AGENTS.md): rules for anyone changing this library.
