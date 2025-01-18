package commands

type Command struct {
	RequireBuild bool
	BuildCmd     []string
	RunCmd       []string
}

var Commands = map[string]Command{
	"C": {
		RequireBuild: true,
		BuildCmd:     []string{"gcc", "-o", "bin/submit", "submit/submit.c"},
		RunCmd:       []string{"bin/submit"},
	},
	"CPP": {
		RequireBuild: true,
		BuildCmd:     []string{"g++", "-o", "bin/submit", "submit/submit.cpp"},
		RunCmd:       []string{"bin/submit"},
	},
	"Java": {
		RequireBuild: true,
		BuildCmd:     []string{"javac", "-d", "bin", "submit/Main.java"},
		RunCmd:       []string{"java", "-cp", "bin", "Main"},
	},
	"Python": {
		RequireBuild: false,
		BuildCmd:     []string{},
		RunCmd:       []string{"python3", "submit/submit.py"},
	},
	"Javascript": {
		RequireBuild: false,
		BuildCmd:     []string{},
		RunCmd:       []string{"node", "submit/submit.js"},
	},
}
