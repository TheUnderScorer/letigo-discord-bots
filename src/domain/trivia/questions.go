package trivia

import (
	"app/logging"
	"app/util"
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

func (question *Question) ID() string {
	return util.Hash(strings.ReplaceAll(question.Question, " ", ""))
}

func (question *Question) Options() []string {
	var options []string
	options = append(options, question.Incorrect...)
	options = append(options, question.Correct)

	return options
}

// ForSpeaking returns the question in a format that can be spoken
func (question *Question) ForSpeaking() string {
	phrase := strings.ReplaceAll(question.Question, "&quot;", "\"")
	phrase = strings.TrimSuffix(phrase, "?")

	return phrase
}

func (question *Question) MarkdownQuestion() string {
	return strings.ReplaceAll(question.Question, "&quot;", "**")
}
