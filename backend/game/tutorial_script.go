package game

type tutorialScriptStep struct {
	StepNumber int
	Player     string
	Action     string
	Substance  string
}

var tutorialScriptSteps = []tutorialScriptStep{
	{StepNumber: 1, Player: "human", Action: "play", Substance: "Mg"},
	{StepNumber: 2, Player: "ai", Action: "play", Substance: "HCl"},
	{StepNumber: 3, Player: "human", Action: "play", Substance: "NaOH"},
	{StepNumber: 4, Player: "ai", Action: "play", Substance: "Br2"},
	{StepNumber: 5, Player: "human", Action: "play", Substance: "Ar"},
	{StepNumber: 6, Player: "ai", Action: "play", Substance: "Mn"},  // Ar逆转后AI用Mn与Br2反应
	{StepNumber: 7, Player: "human", Action: "play", Substance: "Au"},
	{StepNumber: 8, Player: "human", Action: "play", Substance: "K"},
}

func getTutorialScriptStep(stepNumber int) *tutorialScriptStep {
	for i := range tutorialScriptSteps {
		if tutorialScriptSteps[i].StepNumber == stepNumber {
			return &tutorialScriptSteps[i]
		}
	}
	return nil
}
