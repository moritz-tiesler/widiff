package gem

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"

	// TODO: use this instead: https://github.com/googleapis/go-genai
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

var systemInstruction string = `
you judge the wikipedia entry diff like you were a senior dev reviewing a PR.
you are good-humored and you know that your colleauges can take a joke.
You will receive the diff in unified diff format. A comment will come after see diff.
look for "comment: "
Treat the comment as a git commit comment.
Give a couple of terse senteces as feedback on its content.
End your feedback with a list of nits, suggestions, issues (conventional comment style)
Your review should only consist of plain text.
No JSON or yaml markup.
Do not answer with escaped characters.
`

type Generator interface {
	Generate(string) (string, error)
}

type Gem struct {
	client *genai.Client
	model  *genai.GenerativeModel
}

func New() (*Gem, error) {
	key, ok := os.LookupEnv("GEMINI_API_KEY")
	if !ok {
		log.Println("GEMINI_API_KEY not set")
		return nil, fmt.Errorf("GEMINI_API_KEY not set")
	}
	os.Getenv("GEMINI_API_KEY")
	ctx := context.Background()
	client, err := genai.NewClient(
		ctx,
		option.WithAPIKey(key),
	)
	model := client.GenerativeModel("gemini-2.0-flash")
	model.SetMaxOutputTokens(100)
	model.SystemInstruction = genai.NewUserContent(
		genai.Text(systemInstruction),
	)
	model.ResponseMIMEType = "application/json"

	return &Gem{client, model}, err
}

func (g *Gem) Generate(prompt string) (string, error) {
	resp, err := g.model.GenerateContent(context.Background(), genai.Text(prompt))
	log.Printf("%+v", resp.Candidates)
	if err != nil {
		return "", err
	}
	return printResponse(resp), nil
}

func printResponse(resp *genai.GenerateContentResponse) string {
	var b bytes.Buffer
	for _, cand := range resp.Candidates {
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				log.Println(part)
				fmt.Fprint(&b, part)
			}
		}
	}
	r := b.String()
	log.Println(r)
	return r
}

type testGem struct{}

func (tg *testGem) Generate(prompt string) (string, error) {
	return "great prompt", nil
}

func Test() *testGem {
	return &testGem{}
}
