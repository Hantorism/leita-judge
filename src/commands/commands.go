package commands

type Command struct {
	BuildCmd []string
	RunCmd   []string
}

var Commands = map[string]Command{
	"C": {
		BuildCmd: []string{"gcc", "-o", "submit", "submit.c"},
		RunCmd:   []string{"./submit"},
	},
	"CPP": {
		BuildCmd: []string{"g++", "-o", "submit", "submit.cpp"},
		RunCmd:   []string{"./submit"},
	},
}
