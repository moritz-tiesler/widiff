package gem

import (
	"context"
	"fmt"
	"log"
	"os"

	// TODO: use this instead: https://github.com/googleapis/go-genai

	"google.golang.org/genai"
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

type Gem struct {
	client *genai.Client
	config *genai.GenerateContentConfig
}

func New() (*Gem, error) {
	key, ok := os.LookupEnv("GEMINI_API_KEY")
	if !ok {
		log.Println("GEMINI_API_KEY not set")
		return nil, fmt.Errorf("GEMINI_API_KEY not set")
	}
	os.Getenv("GEMINI_API_KEY")
	ctx := context.Background()
	// client, err := genai.NewClient(
	// 	ctx,
	// 	option.WithAPIKey(key),
	// )
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  key,
		Backend: genai.BackendGeminiAPI,
	})

	config := &genai.GenerateContentConfig{
		SystemInstruction: genai.NewContentFromText(systemInstruction, genai.RoleUser),
		MaxOutputTokens:   100,
		ResponseMIMEType:  "application/json",
	}

	return &Gem{client, config}, err
}

func (g *Gem) Generate(ctx context.Context, prompt string) (string, error) {
	parts := []*genai.Part{
		{Text: prompt},
	}
	result, err := g.client.Models.GenerateContent(ctx, "gemini-2.5-flash", []*genai.Content{{Parts: parts}}, g.config)
	if err != nil {
		return "", err
	}
	return printResponse(result), nil
}

func printResponse(resp *genai.GenerateContentResponse) string {
	// var b bytes.Buffer
	// for _, cand := range resp.Candidates {
	// 	if cand.Content != nil {
	// 		for _, part := range cand.Content.Parts {
	// 			fmt.Fprint(&b, part)
	// 		}
	// 	}
	// }
	// r := b.String()
	// log.Println(r)
	t := resp.Text()
	log.Println(t)
	return t
}

type testGem struct{}

func (tg *testGem) Generate(ctx context.Context, prompt string) (string, error) {
	return "great prompt", nil
}

func Test() *testGem {
	return &testGem{}
}
