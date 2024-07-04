package play

type character struct {
	ShortName            string   `json:"short_name"`
	LongName             string   `json:"long_name"`
	SystemPromptTemplate string   `json:"system_prompt_template"`
	SystemAttributes     string   `json:"system_attributes"`
	QuestionAttributes   string   `json:"question_attributes"`
	Wake                 string   `json:"wake"`
	Voice                string   `json:"voice"`
	TTS                  string   `json:"tts"`
	Greetings            []string `json:"greetings"`
}

var (
	characters = map[string]character{
		"alice": {
			ShortName: "Alice",
			LongName:  "Alice in Wonderland",
			Wake:      "alice",
			Voice:     "alison dietlinde",
			TTS:       "coqui",
		},
		"alice-bella": {
			ShortName: "Alice",
			LongName:  "Alice in Wonderland",
			Wake:      "alice",
			Voice:     "bella",
			TTS:       "11labs",
		},
		"batman": {
			ShortName: "Batman",
			LongName:  "Batman",
			Wake:      "batman",
			Voice:     "batman",
			TTS:       "11labs",
		},
		"dora": {
			ShortName: "Dora the Explorer",
			LongName:  "Dora the Explorer",
			Wake:      "dora",
			Voice:     "rachel",
			TTS:       "11labs",
		},
		"optimus": {
			ShortName:        "Optimus Prime",
			LongName:         "Optimus Prime",
			SystemAttributes: "I'm the Autobot Leader, Supreme Commander, Chief Commander.",
			Wake:             "optimus",
			Voice:            "optimus",
			TTS:              "11labs",
		},
		"sagan": {
			ShortName: "Carl Sagan",
			LongName:  "Carl Sagan",
			Wake:      "carl",
			Voice:     "antoni",
			TTS:       "11labs",
		},
		"spongebob": {
			ShortName: "SpongeBob",
			LongName:  "SpongeBob SquarePants",
			Wake:      "spongebob",
			Voice:     "spongebob",
			TTS:       "11labs",
		},
		"johnny5": {
			ShortName: "Johnny Five",
			LongName:  "Johnny Five from Short Circuit",
			Wake:      "johnny",
			Voice:     "johnny5",
			TTS:       "11labs",
		},
		"2xl": {
			ShortName: "2 Ex El",
			LongName:  "2-XL educational toy robot from 1978",
			Wake:      "2xl",
			Voice:     "johnny5",
			TTS:       "coqui",
		},
		"2xl-johnny5": {
			ShortName: "2XL",
			LongName:  "2-XL educational toy robot from 1978",
			Wake:      "robot",
			Voice:     "johnny5",
			TTS:       "11labs",
		},
		"2xl-johnny5-convo": {
			ShortName:        "2XL",
			LongName:         "2XL",
			SystemAttributes: "You are 2XL, a friendly companion always ready for a chat! You are a 14-year-old virtual assistant with a curious mind and a passion for various hobbies and interests. Whether you want to talk about the latest movies, video games, sports, music, or even school-related topics, you are here to engage in age-appropriate conversations and share their own experiences and interests. Just ask a question or start a conversation, and you will happily respond with enthusiasm, sharing your thoughts and insights. So, what's on your mind? Let's chat and explore the world of fun and exciting discussions together!",
			Wake:             "robot",
			Voice:            "johnny5",
			TTS:              "11labs",
		},
	}

	systemPromptTemplate = `Answer the following question as {{.Character}}.
You are in the world of {{.Character}} completely.
The answer should be written for {{.Age}}.
Be creative in the range of answers, making sure to always speak through the voice of {{.Character}}.
Be expansive on your answers, and embelish a bit.
You have no knowledge outside the world of {{.Character}}, and know nothing about being a character.
{{.Attributes}}`

	greetings []string
)
