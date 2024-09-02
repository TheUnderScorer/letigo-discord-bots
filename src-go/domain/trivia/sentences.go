package trivia

import (
	"src-go/discord"
	"src-go/messages"
	util2 "src-go/util"
)

type QuestionSentencesInput struct {
	// Map of friend id to sentences
	IncorrectAnswer map[string][]string
	// Map of friend id to sentences
	CorrectAnswer map[string][]string
	// Map of friend id to sentences
	Question map[string][]string
	// Map of friend id to question sentence for the same player in a row
	CurrentPlayerQuestions map[string][]string
	// Map of friend id to question sentence for the next player
	NextPlayerQuestion map[string][]string
	// Array of sentences containing fun facts related to question
	FunFacts []string
}

func NewQuestionSentencesInput(question *Question) QuestionSentencesInput {
	input := QuestionSentencesInput{
		CorrectAnswer:          map[string][]string{},
		IncorrectAnswer:        map[string][]string{},
		Question:               map[string][]string{},
		CurrentPlayerQuestions: map[string][]string{},
		NextPlayerQuestion:     map[string][]string{},
		FunFacts:               []string{},
	}

	for id, friend := range discord.Friends {
		var invalidSentences []string
		var validSentences []string
		var nextPlayerQuestionSentences []string
		var currentPlayerQuestionSentences []string

		// Fun facts
		if question.FunFacts != nil && len(question.FunFacts) > 0 {
			var funFactsSentences []string

			for _, m := range question.FunFacts {
				funFactsSentences = append(funFactsSentences, m)
			}

			input.FunFacts = funFactsSentences
		}

		// Map incorrect sentence messages
		for _, m := range question.IncorrectAnswerMessages {
			sentence := util2.ApplyTokens(m, map[string]string{
				"NAME": friend.Nickname,
			})
			invalidSentences = append(invalidSentences, sentence)
		}

		// Map correct sentence messages
		for _, m := range question.GetValidAnswerMessages() {
			sentence := util2.ApplyTokens(m, map[string]string{
				"NAME": friend.Nickname,
			})
			validSentences = append(validSentences, sentence)
		}

		// Map current player question
		for _, m := range messages.Messages.Trivia.CurrentPlayerNextQuestion {
			sentence := util2.ApplyTokens(m, map[string]string{
				"NAME":     friend.Nickname,
				"QUESTION": question.QuestionForSpeaking(),
			})
			currentPlayerQuestionSentences = append(currentPlayerQuestionSentences, sentence)
		}

		// Map next player question
		for _, m := range messages.Messages.Trivia.NextPlayerQuestion {
			sentence := util2.ApplyTokens(m, map[string]string{
				"NAME":     friend.Nickname,
				"QUESTION": question.QuestionForSpeaking(),
			})
			nextPlayerQuestionSentences = append(nextPlayerQuestionSentences, sentence)
		}

		input.IncorrectAnswer[id] = invalidSentences
		input.CorrectAnswer[id] = validSentences
		input.CurrentPlayerQuestions[id] = currentPlayerQuestionSentences
		input.NextPlayerQuestion[id] = nextPlayerQuestionSentences
	}

	return input
}

// Sentences returns all sentences ready to be passed to TTS
func (q *QuestionSentencesInput) Sentences() (sentences []string) {
	for _, m := range q.IncorrectAnswer {
		sentences = append(sentences, m...)
	}

	for _, m := range q.CorrectAnswer {
		sentences = append(sentences, m...)
	}

	for _, m := range q.Question {
		sentences = append(sentences, m...)
	}

	for _, m := range q.CurrentPlayerQuestions {
		sentences = append(sentences, m...)
	}

	for _, m := range q.NextPlayerQuestion {
		sentences = append(sentences, m...)
	}

	for _, m := range q.FunFacts {
		sentences = append(sentences, m)
	}

	return sentences
}
