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
		BuildCmd:     []string{"gcc", "submit/temp/Main.c", "-o", "bin/Main", "-O2", "-Wall", "-lm", "-static", "-std=gnu99"},
		RunCmd:       []string{"bin/Main"},
		DeleteCmd:    []string{"rm", "bin/Main"},
	},
	"CPP": {
		RequireBuild: true,
		BuildCmd:     []string{"g++", "submit/temp/Main.cpp", "-o", "bin/Main", "-O2", "-Wall", "-lm", "-static", "-std=gnu++17"},
		RunCmd:       []string{"bin/Main"},
		DeleteCmd:    []string{"rm", "bin/Main"},
	},
	"JAVA": {
		RequireBuild: true,
		BuildCmd:     []string{"javac", "-J-Xms1024m", "-J-Xmx1920m", "-J-Xss512m", "-encoding UTF-8", "-d", "bin", "submit/temp/Main.java"},
		RunCmd:       []string{"java", "-Xms1024m", "-Xmx1920m", "-Xss512m", "-Dfile.encoding=UTF-8", "-XX:+UseSerialGC", "-cp", "bin", "Main"},
		DeleteCmd:    []string{"rm", "bin/Main.class"},
	},
	"PYTHON": {
		RequireBuild: false,
		BuildCmd:     []string{},
		RunCmd:       []string{"python3", "-W", "ignore", "submit/temp/Main.py"},
		DeleteCmd:    []string{},
	},
	"JAVASCRIPT": {
		RequireBuild: false,
		BuildCmd:     []string{},
		RunCmd:       []string{"node", "--stack-size=65536", "submit/temp/Main.js"},
		DeleteCmd:    []string{},
	},
	"GO": {
		RequireBuild: true,
		BuildCmd:     []string{"go", "build", "-o", "bin/Main", "submit/temp/Main.go"},
		RunCmd:       []string{"bin/Main"},
		DeleteCmd:    []string{"rm", "bin/Main"},
	},
	"KOTLIN": {
		RequireBuild: true,
		BuildCmd:     []string{"kotlinc", "-J-Xms1024m", "-J-Xmx1920m", "-J-Xss512m", "-include-runtime", "-d", "bin/Main.jar", "submit/temp/Main.kt"},
		RunCmd:       []string{"java", "-Xms1024m", "-Xmx1920m", "-Xss512m", "-Dfile.encoding=UTF-8", "-XX:+UseSerialGC", "-jar", "bin/Main.jar"},
		DeleteCmd:    []string{"rm", "bin/Main.jar"},
	},
	"SWIFT": {
		RequireBuild: true,
		BuildCmd:     []string{"swiftc", "-o", "bin/Main", "submit/temp/Main.swift"},
		RunCmd:       []string{"bin/Main"},
		DeleteCmd:    []string{"rm", "bin/Main"},
	},
}
