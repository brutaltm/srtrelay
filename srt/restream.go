package srt

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

type CommandWrapper struct {
	Cmd *exec.Cmd
}

func (c *CommandWrapper) runFFmpeg(done chan bool) {
	probesize := 0.1
	time.Sleep(1 * time.Second)
	for {
		var cmd = c.Cmd
		// wrapperCmd.cmd.Stdout = os.Stdout
		// wrapperCmd.cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			fmt.Printf("FFmpeg command exited with error: %s\n", err)
			if <-done {
				fmt.Println("FFmpeg command interrupted -> ending FFmpeg command")
				break
			}
			done <- false
		} else {
			fmt.Println(("FFmpeg command exited without error"))
		}

		newProbesize := probesize + 0.1
		if newProbesize > 2.0 || cmd.ProcessState.UserTime()+cmd.ProcessState.SystemTime() > 10*time.Second {
			newProbesize = probesize
		}

		cmd.Args[2] = strings.Replace(cmd.Args[2], fmt.Sprintf("-probesize %v", probesize), fmt.Sprintf("-probesize %v", newProbesize), 1)
		fmt.Println("Restarting FFmpeg command with probesize:", probesize, "M -> ", newProbesize, "M")
		fmt.Println("New ffmpeg command:", cmd.Args[2])
		probesize = newProbesize
		c.Cmd = exec.Command("sh", "-c", cmd.Args[2])
		continue
	}
}

func (c *CommandWrapper) close(done chan bool) {
	cmd := c.Cmd
	if len(done) > 0 {
		<-done
	}
	done <- true
	err2 := cmd.Process.Signal(os.Interrupt)
	if err2 != nil {
		log.Fatalf("Failed to kill process with PID %d: %s\n", cmd.Process.Pid, err2)
	}
	fmt.Println("FFmpeg command with PID", cmd.Process.Pid, "was interrupted")
}
