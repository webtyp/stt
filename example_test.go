package stt_test

import (
	"testing"

	"webtyp.com/audio"
	"webtyp.com/context"
	"webtyp.com/fmt"
	"webtyp.com/stt"
)

type fixedTranscriber struct {
	text string
}

func (t fixedTranscriber) Transcribe(ctx *context.Context, speech audio.PCM) (stt.Transcript, error) {
	if speech.SampleRate != 16000 {
		return stt.Transcript{}, fmt.Err("stt: expected 16000 Hz")
	}
	return stt.Transcript{
		Text:     t.text,
		Language: "es",
	}, nil
}

type wordTranscriber struct{}

func (wt wordTranscriber) NewStream(ctx *context.Context) (stt.Stream, error) {
	return &wordStream{
		words: []string{"reserva", "una", "hora"},
	}, nil
}

type wordStream struct {
	words  []string
	index  int
	closed bool
}

func (s *wordStream) Write(speech audio.PCM) (stt.Transcript, error) {
	if s.closed {
		return stt.Transcript{}, fmt.Err("stt: stream closed")
	}
	if s.index < len(s.words) {
		s.index++
	}
	return stt.Transcript{
		Text:     joinWords(s.words[:s.index]),
		Language: "es",
	}, nil
}

func (s *wordStream) Close() (stt.Transcript, error) {
	if s.closed {
		return stt.Transcript{}, fmt.Err("stt: stream closed")
	}
	s.closed = true
	return stt.Transcript{
		Text:     joinWords(s.words),
		Language: "es",
	}, nil
}

func joinWords(words []string) string {
	res := ""
	for i, w := range words {
		if i > 0 {
			res += " "
		}
		res += w
	}
	return res
}

func Example_dictation() {
	ctx := context.Background()
	tr := fixedTranscriber{text: "reserva una hora"}
	speech := audio.PCM{
		SampleRate: 16000,
		Samples:    make([]float32, 16000),
	}

	res, err := tr.Transcribe(ctx, speech)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	fmt.Println(res.Text)
	// Output:
	// reserva una hora
}

func TestTranscribe_RejectsWrongRate(t *testing.T) {
	ctx := context.Background()
	tr := fixedTranscriber{text: "reserva una hora"}
	speech := audio.PCM{
		SampleRate: 48000,
		Samples:    make([]float32, 48000),
	}

	_, err := tr.Transcribe(ctx, speech)
	if err == nil {
		t.Fatal("expected error for wrong sample rate, got nil")
	}
}

func TestStream_PartialThenFinal(t *testing.T) {
	var _ stt.StreamTranscriber = wordTranscriber{}

	ctx := context.Background()
	stTr := wordTranscriber{}
	stream, err := stTr.NewStream(ctx)
	if err != nil {
		t.Fatalf("unexpected error creating stream: %v", err)
	}

	speech := audio.PCM{SampleRate: 16000}

	res1, err := stream.Write(speech)
	if err != nil {
		t.Fatalf("Write 1 error: %v", err)
	}
	if res1.Text != "reserva" {
		t.Errorf("Write 1 text = %q, want %q", res1.Text, "reserva")
	}

	res2, err := stream.Write(speech)
	if err != nil {
		t.Fatalf("Write 2 error: %v", err)
	}
	if res2.Text != "reserva una" {
		t.Errorf("Write 2 text = %q, want %q", res2.Text, "reserva una")
	}

	res3, err := stream.Write(speech)
	if err != nil {
		t.Fatalf("Write 3 error: %v", err)
	}
	if res3.Text != "reserva una hora" {
		t.Errorf("Write 3 text = %q, want %q", res3.Text, "reserva una hora")
	}

	resClose, err := stream.Close()
	if err != nil {
		t.Fatalf("Close error: %v", err)
	}
	if resClose.Text != "reserva una hora" {
		t.Errorf("Close text = %q, want %q", resClose.Text, "reserva una hora")
	}

	_, err = stream.Write(speech)
	if err == nil {
		t.Fatal("expected error writing to closed stream, got nil")
	}
}
