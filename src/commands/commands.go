package commands

type Command struct {
	BuildCmd []string
	RunCmd   []string
}

var Commands = map[string]Command{
	"C": {
		BuildCmd: []string{"gcc", "-o", "bin/submit", "submit/submit.c"},
		RunCmd:   []string{"bin/submit"},
	},
	"CPP": {
		BuildCmd: []string{"g++", "-o", "bin/submit", "submit/submit.cpp"},
		RunCmd:   []string{"bin/submit"},
	},
	"Java": {
		BuildCmd: []string{"javac", "-d", "bin", "submit/Main.java"},
		RunCmd:   []string{"java", "-cp", "bin", "Main"},
	},
}
