package commands

type Command struct {
	BuildCmd  []string
	RunCmd    []string
	DeleteCmd []string
}

var Commands = map[string]Command{
	"C": {
		BuildCmd:  []string{"gcc", "{JUDGE_TYPE}/{SUBMIT_ID}/Main.c", "-o", "{JUDGE_TYPE}/{SUBMIT_ID}/Main", "-O2", "-Wall", "-lm", "-static", "-std=gnu99"},
		RunCmd:    []string{"{JUDGE_TYPE}/{SUBMIT_ID}/Main"},
		DeleteCmd: []string{"rm", "{JUDGE_TYPE}/{SUBMIT_ID}/Main"},
	},
	"CPP": {
		BuildCmd:  []string{"g++", "{JUDGE_TYPE}/{SUBMIT_ID}/Main.cpp", "-o", "{JUDGE_TYPE}/{SUBMIT_ID}/Main", "-O2", "-Wall", "-lm", "-static", "-std=gnu++17"},
		RunCmd:    []string{"{JUDGE_TYPE}/{SUBMIT_ID}/Main"},
		DeleteCmd: []string{"rm", "{JUDGE_TYPE}/{SUBMIT_ID}/Main"},
	},
	"JAVA": {
		BuildCmd:  []string{"javac", "-J-Xms1024m", "-J-Xmx1920m", "-J-Xss512m", "-encoding UTF-8", "-d", "bin", "{JUDGE_TYPE}/{SUBMIT_ID}/Main.java"},
		RunCmd:    []string{"java", "-Xms1024m", "-Xmx1920m", "-Xss512m", "-Dfile.encoding=UTF-8", "-XX:+UseSerialGC", "-cp", "bin", "Main"},
		DeleteCmd: []string{"rm", "{JUDGE_TYPE}/{SUBMIT_ID}/Main.class"},
	},
	"PYTHON": {
		BuildCmd:  []string{},
		RunCmd: []string{"python3", "-W", "ignore", "{JUDGE_TYPE}/{SUBMIT_ID}/Main.py"},
		DeleteCmd: []string{},
	},
	"JAVASCRIPT": {
		BuildCmd:  []string{},
		RunCmd: []string{"node", "--stack-size=65536", "{JUDGE_TYPE}/{SUBMIT_ID}/Main.js"},
		DeleteCmd: []string{},
	},
	"GO": {
		BuildCmd:  []string{"go", "build", "-o", "{JUDGE_TYPE}/{SUBMIT_ID}/Main", "{JUDGE_TYPE}/{SUBMIT_ID}/Main.go"},
		RunCmd:    []string{"{JUDGE_TYPE}/{SUBMIT_ID}/Main"},
		DeleteCmd: []string{"rm", "{JUDGE_TYPE}/{SUBMIT_ID}/Main"},
	},
	"KOTLIN": {
		BuildCmd:  []string{"kotlinc", "-J-Xms1024m", "-J-Xmx1920m", "-J-Xss512m", "-include-runtime", "-d", "{JUDGE_TYPE}/{SUBMIT_ID}/Main.jar", "{JUDGE_TYPE}/{SUBMIT_ID}/Main.kt"},
		RunCmd:    []string{"java", "-Xms1024m", "-Xmx1920m", "-Xss512m", "-Dfile.encoding=UTF-8", "-XX:+UseSerialGC", "-jar", "{JUDGE_TYPE}/{SUBMIT_ID}/Main.jar"},
		DeleteCmd: []string{"rm", "{JUDGE_TYPE}/{SUBMIT_ID}/Main.jar"},
	},
	"SWIFT": {
		BuildCmd:  []string{"swiftc", "-O", "-o", "{JUDGE_TYPE}/{SUBMIT_ID}/Main", "{JUDGE_TYPE}/{SUBMIT_ID}/Main.swift"},
		RunCmd:    []string{"{JUDGE_TYPE}/{SUBMIT_ID}/Main"},
		DeleteCmd: []string{"rm", "{JUDGE_TYPE}/{SUBMIT_ID}/Main"},
	},
}
