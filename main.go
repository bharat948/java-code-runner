package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os/exec"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
)

type CodeSubmission struct {
	Code      string `json:"code"`
	Language  string `json:"language"`
	TestInput string `json:"test_input"`
}

type ExecutionResult struct {
	Output      string `json:"output"`
	Error       string `json:"error"`
	MemoryUsage string `json:"memory_usage"`
	CPUUsage    string `json:"cpu_usage"`
}

func main() {
	r := gin.Default()

	r.POST("/run", func(c *gin.Context) {
		var submission CodeSubmission
		if err := c.ShouldBindJSON(&submission); err != nil {
			c.JSON(400, gin.H{"error": "Invalid input"})
			return
		}

		if submission.Language != "java" {
			c.JSON(400, gin.H{"error": "Only Java language is supported"})
			return
		}

		// Save the submitted code to a file
		err := ioutil.WriteFile("Main.java", []byte(submission.Code), 0644)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to save code"})
			return
		}

		// Compile the Java code
		compileCmd := exec.Command("javac", "-d", "./", "Main.java")
		var compileStderr bytes.Buffer
		compileCmd.Stderr = &compileStderr
		err = compileCmd.Run()
		if err != nil {
			c.JSON(500, ExecutionResult{
				Output:      "",
				Error:       fmt.Sprintf("Compilation failed: %s", compileStderr.String()),
				MemoryUsage: "",
				CPUUsage:    "",
			})
			return
		}

		// Run the compiled Java code
		runCmd := exec.Command("java", "-cp", "./", "Main")
		runCmd.Stdin = bytes.NewReader([]byte(submission.TestInput))
		var runStdout, runStderr bytes.Buffer
		runCmd.Stdout = &runStdout
		runCmd.Stderr = &runStderr

		startTime := time.Now()
		err = runCmd.Run()
		executionTime := time.Since(startTime)

		// Collect runtime statistics (simplified for illustration)
		memStats := &runtime.MemStats{}
		runtime.ReadMemStats(memStats)

		if err != nil {
			c.JSON(500, ExecutionResult{
				Output:      "",
				Error:       fmt.Sprintf("Runtime error: %s\nStderr: %s", err, runStderr.String()),
				MemoryUsage: fmt.Sprintf("%d KB", memStats.Alloc/1024),
				CPUUsage:    fmt.Sprintf("%v ms", executionTime.Milliseconds()),
			})
			return
		}

		c.JSON(200, ExecutionResult{
			Output:      runStdout.String(),
			Error:       "",
			MemoryUsage: fmt.Sprintf("%d KB", memStats.Alloc/1024),
			CPUUsage:    fmt.Sprintf("%v ms", executionTime.Milliseconds()),
		})
	})

	r.Run(":8080")
}
