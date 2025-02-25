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
		BuildCmd:     []string{"gcc", "submit/{SUBMITID}/Main.c", "-o", "submit/{SUBMITID}/Main", "-O2", "-Wall", "-lm", "-static", "-std=gnu99"},
		RunCmd:       []string{"submit/{SUBMITID}/Main"},
		DeleteCmd:    []string{"rm", "submit/{SUBMITID}/Main"},
	},
	"CPP": {
		RequireBuild: true,
		BuildCmd:     []string{"g++", "submit/{SUBMITID}/Main.cpp", "-o", "submit/{SUBMITID}/Main", "-O2", "-Wall", "-lm", "-static", "-std=gnu++17"},
		RunCmd:       []string{"submit/{SUBMITID}/Main"},
		DeleteCmd:    []string{"rm", "submit/{SUBMITID}/Main"},
	},
	"JAVA": {
		RequireBuild: true,
		BuildCmd:     []string{"javac", "-J-Xms1024m", "-J-Xmx1920m", "-J-Xss512m", "-encoding UTF-8", "-d", "bin", "submit/{SUBMITID}/Main.java"},
		RunCmd:       []string{"java", "-Xms1024m", "-Xmx1920m", "-Xss512m", "-Dfile.encoding=UTF-8", "-XX:+UseSerialGC", "-cp", "bin", "Main"},
		DeleteCmd:    []string{"rm", "submit/{SUBMITID}/Main.class"},
	},
	"PYTHON": {
		RequireBuild: false,
		BuildCmd:     []string{},
		RunCmd:       []string{"python3", "-W", "ignore", "submit/{SUBMITID}/Main.py"},
		DeleteCmd:    []string{},
	},
	"JAVASCRIPT": {
		RequireBuild: false,
		BuildCmd:     []string{},
		RunCmd:       []string{"node", "--stack-size=65536", "submit/{SUBMITID}/Main.js"},
		DeleteCmd:    []string{},
	},
	"GO": {
		RequireBuild: true,
		BuildCmd:     []string{"go", "build", "-o", "submit/{SUBMITID}/Main", "submit/{SUBMITID}/Main.go"},
		RunCmd:       []string{"submit/{SUBMITID}/Main"},
		DeleteCmd:    []string{"rm", "submit/{SUBMITID}/Main"},
	},
	"KOTLIN": {
		RequireBuild: true,
		BuildCmd:     []string{"kotlinc", "-J-Xms1024m", "-J-Xmx1920m", "-J-Xss512m", "-include-runtime", "-d", "submit/{SUBMITID}/Main.jar", "submit/{SUBMITID}/Main.kt"},
		RunCmd:       []string{"java", "-Xms1024m", "-Xmx1920m", "-Xss512m", "-Dfile.encoding=UTF-8", "-XX:+UseSerialGC", "-jar", "submit/{SUBMITID}/Main.jar"},
		DeleteCmd:    []string{"rm", "submit/{SUBMITID}/Main.jar"},
	},
	"SWIFT": {
		RequireBuild: true,
		BuildCmd:     []string{"swiftc", "-O", "-o", "submit/{SUBMITID}/Main", "submit/{SUBMITID}/Main.swift"},
		RunCmd:       []string{"submit/{SUBMITID}/Main"},
		DeleteCmd:    []string{"rm", "submit/{SUBMITID}/Main"},
	},
}
