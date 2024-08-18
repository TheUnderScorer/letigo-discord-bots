package trivia

import (
	_ "embed"
	"encoding/json"
	"strings"
)

//go:embed questions.json
var questions []byte

type QuestionType string

var (
	MultipleChoice QuestionType = "multiple"
	TrueFalse      QuestionType = "boolean"
)

type QuestionDifficulty string

var (
	Easy   QuestionDifficulty = "easy"
	Medium QuestionDifficulty = "medium"
	Hard   QuestionDifficulty = "hard"
)

type Question struct {
	Type                    QuestionType       `json:"type"`
	Difficulty              QuestionDifficulty `json:"difficulty"`
	Category                string             `json:"category"`
	Question                string             `json:"question"`
	Correct                 string             `json:"correct_answer"`
	Incorrect               []string           `json:"incorrect_answers"`
	IncorrectAnswerMessages []string           `json:"incorrect_anwser_messages"`
	CorrectAnswerMessages   []string           `json:"correct_answer_messages"`
	FunFacts                []string           `json:"fun_facts"`
}

func GetQuestions() []Question {
	var result []Question
	_ = json.Unmarshal(questions, &result)
	return result
}

func (q *Question) Options() []string {
	var options []string
	options = append(options, q.Incorrect...)
	options = append(options, q.Correct)

	return options
}

func (q *Question) QuestionForSpeaking() string {
	return strings.ReplaceAll(q.Question, "&quot;", "\"")
}

func (q *Question) MarkdownQuestion() string {
	return strings.ReplaceAll(q.Question, "&quot;", "**")
}
