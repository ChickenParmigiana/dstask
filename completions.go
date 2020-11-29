package dstask

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

// Completions ...
func Completions(conf Config, args []string, ctx CmdLine) {
	// given the entire user's command line arguments as the arguments for
	// this cmd, suggest possible candidates for the last arg.
	// see the relevant shell completion bindings in this repository for
	// integration. Note there are various idiosyncrasies with bash
	// involving arg separation.
	var completions []string
	var originalArgs []string
	var prefix string

	// drop dstask _completions dstask to allow parsing what is on actual
	// prompt
	if len(args) > 3 {
		originalArgs = args[3:]
	}

	// args are dstask _completions <user command line>
	// parse command line as normal to set rules
	cmdLine := ParseCmdLine(originalArgs...)

	// No command and OK to specify command (to run or help)
	// Note that techically we should only specify commands as available
	// completions if the last partial argument is a command substring.
	// However, this is unnecessary as a general substring filter is used at
	// the end of the func.
	// This is exhaustive but the clearest way, IMO.
	if len(cmdLine.AntiProjects) == 0 &&
		cmdLine.Project == "" &&
		len(cmdLine.Tags) == 0 &&
		len(cmdLine.AntiTags) == 0 &&
		cmdLine.Priority == "" &&
		cmdLine.Template == 0 &&
		!cmdLine.IgnoreContext &&
		(cmdLine.Cmd == CMD_HELP || cmdLine.Cmd == "") {
		for _, cmd := range ALL_CMDS {
			if !strings.HasPrefix(cmd, "_") {
				completions = append(completions, cmd)
			}
		}
	}

	if StrSliceContains([]string{
		"",
		CMD_NEXT,
		CMD_ADD,
		CMD_REMOVE,
		CMD_LOG,
		CMD_START,
		CMD_STOP,
		CMD_DONE,
		CMD_RESOLVE,
		CMD_CONTEXT,
		CMD_MODIFY,
		CMD_SHOW_NEXT,
		CMD_SHOW_PROJECTS,
		CMD_SHOW_ACTIVE,
		CMD_SHOW_PAUSED,
		CMD_SHOW_OPEN,
		CMD_SHOW_RESOLVED,
		CMD_SHOW_TEMPLATES,
	}, cmdLine.Cmd) {
		ts, err := NewTaskSet(
			conf.Repo, conf.IDsFile, conf.StateFile,
			WithStatuses(NON_RESOLVED_STATUSES...),
		)
		if err != nil {
			log.Printf("completions script error: %v\n", err)
			return

		}
		// limit completions to available context, but not if the user is
		// trying to change context, context ignore is on, or modify
		// command is being completed
		if !cmdLine.IgnoreContext &&
			cmdLine.Cmd != CMD_CONTEXT &&
			cmdLine.Cmd != CMD_MODIFY {
			ts.Filter(ctx)
		}

		// templates
		if cmdLine.Cmd == CMD_ADD {
			for _, task := range ts.Tasks() {
				if task.Status == STATUS_TEMPLATE {
					completions = append(completions, "template:"+strconv.Itoa(task.ID))
				}
			}
		}

		// priorities
		completions = append(completions, PRIORITY_CRITICAL)
		completions = append(completions, PRIORITY_HIGH)
		completions = append(completions, PRIORITY_NORMAL)
		completions = append(completions, PRIORITY_LOW)

		// projects
		for project := range ts.GetProjects() {
			completions = append(completions, "project:"+project)
			completions = append(completions, "-project:"+project)
		}

		// tags
		for tag := range ts.GetTags() {
			completions = append(completions, "+"+tag)
			completions = append(completions, "-"+tag)
		}
	}

	if len(originalArgs) > 0 {
		prefix = originalArgs[len(originalArgs)-1]
	}

	for _, completion := range completions {
		if strings.HasPrefix(completion, prefix) && !StrSliceContains(originalArgs, completion) {
			fmt.Println(completion)
		}
	}
}
