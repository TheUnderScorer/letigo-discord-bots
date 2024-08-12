package opentdb

import (
	"encoding/json"
	"net/http"
	"src-go/domain/trivia"
)

type Question struct {
	Type             trivia.QuestionType       `json:"type"`
	Difficulty       trivia.QuestionDifficulty `json:"difficulty"`
	Category         string                    `json:"category"`
	Question         string                    `json:"question"`
	CorrectAnswer    string                    `json:"correct_answer"`
	IncorrectAnswers []string                  `json:"incorrect_answers"`
}

func (q *Question) ToTrivia() trivia.Question {
	return trivia.Question{
		Question:                q.Question,
		Incorrect:               q.IncorrectAnswers,
		Correct:                 q.CorrectAnswer,
		Type:                    q.Type,
		Difficulty:              q.Difficulty,
		Category:                q.Category,
		IncorrectAnswerMessages: make([]string, 0),
		CorrectAnswerMessages:   make([]string, 0),
	}
}

type Response struct {
	ResponseCode int        `json:"response_code"`
	Results      []Question `json:"results"`
}

var client = http.Client{}

const url = "https://opentdb.com/api.php?amount=15"

func GetQuestions() ([]Question, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var r Response
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}

	return r.Results, nil
}
