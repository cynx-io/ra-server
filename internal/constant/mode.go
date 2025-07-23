package constant

type ModeType string

const (
	// input: answer, output: answer
	ModeTypeWordle ModeType = "WORDLE"
	ModeTypeSudoku ModeType = "SUDOKU"

	// input: answer string, output: iscorrect
	ModeTypeHangman ModeType = "HANGMAN"

	// Tbc
	ModeTypeMemory             ModeType = "MEMORY"
	ModeTypeQuiz               ModeType = "QUIZ"
	ModeTypeCrossword          ModeType = "CROSSWORD"
	ModeTypePuzzle             ModeType = "PUZZLE"
	ModeTypeTrivia             ModeType = "TRIVIA"
	ModeTypeFlashcards         ModeType = "FLASHCARDS"
	ModeTypeMatching           ModeType = "MATCHING"
	ModeTypeFillInTheBlank     ModeType = "FILL_IN_THE_BLANK"
	ModeTypeMultipleChoice     ModeType = "MULTIPLE_CHOICE"
	ModeTypeTrueFalse          ModeType = "TRUE_FALSE"
	ModeTypeSorting            ModeType = "SORTING"
	ModeTypeSequence           ModeType = "SEQUENCE"
	ModeTypeWordSearch         ModeType = "WORD_SEARCH"
	ModeTypeAnagram            ModeType = "ANAGRAM"
	ModeTypeRiddles            ModeType = "RIDDLES"
	ModeTypeLogicPuzzle        ModeType = "LOGIC_PUZZLE"
	ModeTypeMathPuzzle         ModeType = "MATH_PUZZLE"
	ModeTypeVisualPuzzle       ModeType = "VISUAL_PUZZLE"
	ModeTypeAudioPuzzle        ModeType = "AUDIO_PUZZLE"
	ModeTypeCodePuzzle         ModeType = "CODE_PUZZLE"
	ModeTypeEscapeRoom         ModeType = "ESCAPE_ROOM"
	ModeTypeScavengerHunt      ModeType = "SCAVENGER_HUNT"
	ModeTypeStoryPuzzle        ModeType = "STORY_PUZZLE"
	ModeTypeWordAssociation    ModeType = "WORD_ASSOCIATION"
	ModeTypeNumberPuzzle       ModeType = "NUMBER_PUZZLE"
	ModeTypePatternRecognition ModeType = "PATTERN_RECOGNITION"
	ModeTypeTriviaChallenge    ModeType = "TRIVIA_CHALLENGE"
	ModeTypeFlashQuiz          ModeType = "FLASH_QUIZ"
	ModeTypeInteractiveStory   ModeType = "INTERACTIVE_STORY"
	ModeTypeCreativeWriting    ModeType = "CREATIVE_WRITING"
)
