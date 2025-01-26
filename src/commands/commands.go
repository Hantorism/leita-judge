package commands

type Command struct {
	RequireBuild bool
	BuildCmd     []string
	RunCmd       []string
	DeleteCmd    []string
}

var Commands = map[string]Command{
	"C": {
		RequireBuild: true,
		BuildCmd:     []string{"gcc", "-o", "bin/Main", "submit/Main.c"},
		RunCmd:       []string{"bin/Main"},
		DeleteCmd:    []string{"rm", "bin/Main"},
	},
	"CPP": {
		RequireBuild: true,
		BuildCmd:     []string{"g++", "-o", "bin/Main", "submit/Main.cpp"},
		RunCmd:       []string{"bin/Main"},
		DeleteCmd:    []string{"rm", "bin/Main"},
	},
	"JAVA": {
		RequireBuild: true,
		BuildCmd:     []string{"javac", "-d", "bin", "submit/Main.java"},
		RunCmd:       []string{"java", "-cp", "bin", "Main"},
		DeleteCmd:    []string{"rm", "bin/Main.class"},
	},
	"PYTHON": {
		RequireBuild: false,
		BuildCmd:     []string{},
		RunCmd:       []string{"python3", "submit/Main.py"},
		DeleteCmd:    []string{},
	},
	"JAVASCRIPT": {
		RequireBuild: false,
		BuildCmd:     []string{},
		RunCmd:       []string{"node", "submit/Main.js"},
		DeleteCmd:    []string{},
	},
	"GO": {
		RequireBuild: true,
		BuildCmd:     []string{"go", "build", "-o", "bin/Main", "submit/Main.go"},
		RunCmd:       []string{"bin/Main"},
		DeleteCmd:    []string{"rm", "bin/Main"},
	},
	"KOTLIN": {
		RequireBuild: true,
		BuildCmd:     []string{"kotlinc", "submit/Main.kt", "-include-runtime", "-d", "bin/Main.jar"},
		RunCmd:       []string{"java", "-jar", "bin/Main.jar"},
		DeleteCmd:    []string{"rm", "bin/Main.jar"},
	},
	"SWIFT": {
		RequireBuild: true,
		BuildCmd:     []string{"swiftc", "-o", "bin/Main", "submit/Main.swift"},
		RunCmd:       []string{"bin/Main"},
		DeleteCmd:    []string{"rm", "bin/Main"},
	},
}
