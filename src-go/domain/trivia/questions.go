package trivia

import (
	"src-go/logging"
	"src-go/messages"
	"src-go/util"
	"strings"
)

var log = logging.Get().Named("questions")

type QuestionType string

var (
	MultipleChoice QuestionType = "multiple"
	TrueFalse      QuestionType = "boolean"
)

const True = "Prawda"
const False = "Fałsz"

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

func (q *Question) ID() string {
	return util.Hash(strings.ReplaceAll(q.Question, " ", ""))
}

func (q *Question) Options() []string {
	var options []string
	options = append(options, q.Incorrect...)
	options = append(options, q.Correct)

	return options
}

func (q *Question) GetValidAnswerMessages() (m []string) {
	if q.Type == TrueFalse {
		m = messages.Messages.Trivia.ValidAnswer.Boolean
	} else {
		m = messages.Messages.Trivia.ValidAnswer.Multiple
	}

	return m
}

// ForSpeaking returns the question in a format that can be spoken
func (q *Question) ForSpeaking() string {
	question := strings.ReplaceAll(q.Question, "&quot;", "\"")
	question = strings.TrimSuffix(question, "?")

	return question
}

func (q *Question) MarkdownQuestion() string {
	return strings.ReplaceAll(q.Question, "&quot;", "**")
}
