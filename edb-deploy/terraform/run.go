package terraform

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

var templateLocation = ""

func RunTerraform(projectName string, project map[string]interface{}, arguements map[string]interface{}, variables map[string]interface{}, fileName string, customTemplateLocation *string) error {
	// Retrieve from Environment variable debugging setting
	verbose = getDebuggingStateFromOS()

	if verbose {
		fmt.Println("--- Debugging - terraform -> run.go - RunTerraform:")
		fmt.Println("Starting...")
		fmt.Println("---")
	}

	if customTemplateLocation != nil {
		templateLocation = *customTemplateLocation
	} else {
		path, err := os.Getwd()
		if err != nil {
			log.Println(err)
		}

		splitPath := strings.Split(path, "/")

		if len(splitPath) > 0 {
			splitPath = splitPath[:len(splitPath)-1]
		}

		splitPath = append(splitPath, "terraform")
		splitPath = append(splitPath, fileName)

		templateLocation = strings.Join(splitPath, "/")
	}

	project["project_name"] = projectName

	setHardCodedVariables(project, variables)
	setMappedVariables(project, variables)
	setVariableAndTagNames(projectName)

	if verbose {
		fmt.Println("--- Debugging - terraform -> run.go - RunTerraform:")
		fmt.Println("Calling: 'pre_run_checks'...")
		fmt.Println("project")
		fmt.Println(project)
		fmt.Println("---")
	}

	if arguements["pre_run_checks"] != nil {
		preRunChecks := arguements["pre_run_checks"].(map[string]interface{})

		if verbose {
			fmt.Println("--- Debugging - terraform -> run.go - RunTerraform:")
			fmt.Println("preRunChecks")
			fmt.Println(preRunChecks)
			fmt.Println("len of preRunChecks")
			fmt.Println(len(preRunChecks))
			fmt.Println("---")
		}

		for i := 0; i < len(preRunChecks); i++ {
			iString := strconv.Itoa(i)

			if verbose {
				fmt.Println("iString")
				fmt.Println(iString)
			}

			check := preRunChecks[iString].(map[string]interface{})

			if verbose {
				fmt.Println("check")
				fmt.Println(check)
			}

			output, _ := preCheck(check, project)

			if verbose {
				fmt.Println("project")
				fmt.Println(project)
				fmt.Println("output")
				fmt.Println(output)
			}

			if check["output"] != nil {
				project[check["output"].(string)] = output
			}
		}
	}

	if verbose {
		fmt.Println("--- Debugging - terraform -> run.go - RunTerraform:")
		fmt.Println("Returned from Calling: 'pre_run_checks'...")
		fmt.Println("---")
	}

	terraformWorkspace(projectName)

	cmd := exec.Command("terraform", "init")
	cmd.Dir = templateLocation

	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("%s\n", stdoutStderr)
		log.Fatal(err)
	}

	if arguements["terraform_build"] != nil {
		terraformBuild := arguements["terraform_build"].(map[string]interface{})
		argSlice := terraformBuild["variables"].([]interface{})
		terraformApply(argSlice, project)
	}

	if arguements["post_run_checks"] != nil {
		postRunChecks := arguements["post_run_checks"].(map[string]interface{})

		for i := 0; i < len(postRunChecks); i++ {
			iString := strconv.Itoa(i)
			check := postRunChecks[iString].(map[string]interface{})

			output, _ := preCheck(check, project)
			if check["output"] != nil {
				project[check["output"].(string)] = output
			}
		}
	}

	if verbose {
		fmt.Println("--- Debugging - terraform -> run.go - RunTerraform:")
		fmt.Println("Leaving...")
		fmt.Println("---")
	}

	return nil
}

func setVariableAndTagNames(projectName string) error {
	tagsInput := "/tags.tf.template"
	tagsOutput := "/tags.tf"

	variablesInput := "/variables.tf.template"
	variablesOutput := "/variables.tf"

	textToReplace := "PROJECT_NAME"

	if verbose {
		fmt.Println("--- Debugging - terraform -> run.go - setVariableAndTagNames:")
		fmt.Println("Starting...")
		fmt.Println("---")
	}

	tagTemplate, err := ioutil.ReadFile(fmt.Sprintf("%s%s", templateLocation, tagsInput))
	if err != nil {
		log.Fatal(err)
	}

	tagsReplaced := bytes.ReplaceAll(tagTemplate, []byte(textToReplace), []byte(projectName))

	err = ioutil.WriteFile(fmt.Sprintf("%s%s", templateLocation, tagsOutput), tagsReplaced, 0644)
	if err != nil {
		log.Fatal(err)
	}

	variableTemplate, err := ioutil.ReadFile(fmt.Sprintf("%s%s", templateLocation, variablesInput))
	if err != nil {
		log.Fatal(err)
	}

	variablesReplaced := bytes.ReplaceAll(variableTemplate, []byte(textToReplace), []byte(projectName))

	err = ioutil.WriteFile(fmt.Sprintf("%s%s", templateLocation, variablesOutput), variablesReplaced, 0644)
	if err != nil {
		log.Fatal(err)
	}

	if verbose {
		fmt.Println("--- Debugging - terraform -> run.go - setVariableAndTagNames:")
		fmt.Println("Leaving...")
		fmt.Println("---")
	}

	return nil
}

func setHardCodedVariables(project map[string]interface{}, variables map[string]interface{}) error {
	if verbose {
		fmt.Println("--- Debugging - terraform -> run.go - setHardCodedVariables:")
		fmt.Println("Starting...")
		fmt.Println("---")
	}

	if variables != nil {
		hardCoded := variables["hard"].(map[string]interface{})

		for variable, value := range hardCoded {
			project[variable] = value
		}
	}

	if verbose {
		fmt.Println("--- Debugging - terraform -> run.go - setHardCodedVariables:")
		fmt.Println("Leaving...")
		fmt.Println("---")
	}

	return nil
}

func setMappedVariables(project map[string]interface{}, variables map[string]interface{}) error {
	if verbose {
		fmt.Println("--- Debugging - terraform -> run.go - setMappedVariables:")
		fmt.Println("DEBUG")
		fmt.Println(verbose)
		fmt.Println("project")
		fmt.Println(project)
		fmt.Println("---")
	}

	if variables != nil {
		maps := variables["maps"].(map[string]interface{})

		if verbose {
			fmt.Println("--- Debugging:")
			fmt.Println("project")
			fmt.Println(project)
			fmt.Println("maps")
			fmt.Println(maps)
			fmt.Println("---")
		}

		for input, mapArray := range maps {
			mArr := mapArray.(map[string]interface{})
			for _, mMap := range mArr {
				m := mMap.(map[string]interface{})
				actualMap := m["map"].(map[string]interface{})
				out := ""

				if verbose {
					fmt.Println("--- Debugging:")
					fmt.Println("Interface for: m")
					fmt.Println(m)
					fmt.Println("project")
					fmt.Println(project)
					fmt.Println("input")
					fmt.Println(input)
				}

				if m["type"] == "starts-with" {
					for criteria, value := range actualMap {
						if strings.HasPrefix(project[input].(string), criteria) {
							out = value.(string)
						}
					}
				} else {
					val := project[input].(string)
					out = actualMap[val].(string)
				}

				project[m["output"].(string)] = out
				if verbose {
					fmt.Println("out")
					fmt.Println(out)
					fmt.Println("---")
				}
			}
		}
	}

	return nil
}

func preCheck(check map[string]interface{}, project map[string]interface{}) (string, error) {
	if check["log_text"] != nil {
		log.Printf(check["log_text"].(string))
	}

	output := ""

	if check["command"] != nil {
		command := check["command"].(string)
		if check["variables"] != nil {
			variables := check["variables"].(map[string]interface{})

			for i := 0; i < len(variables); i++ {
				iString := strconv.Itoa(i)
				variable := variables[iString].(string)

				if project[variable] != nil {
					value := strings.ReplaceAll(project[variable].(string), " ", "|||")
					command = strings.Replace(command, "%s", value, 1)
				}
			}
		}

		splitCommand := strings.Split(command, " ")

		for i, c := range splitCommand {
			splitCommand[i] = strings.ReplaceAll(c, "|||", " ")
		}

		comm := exec.Command(splitCommand[0], splitCommand[1:len(splitCommand)]...)

		stdoutStderr, err := comm.CombinedOutput()
		if err != nil {
			log.Fatal(err)
		}

		output = strings.ReplaceAll(string(stdoutStderr), "\n", "")
	}

	return output, nil
}

func terraformWorkspace(projectName string) error {
	log.Printf("Checking Projects in terraform")
	comm := exec.Command("terraform", "workspace", "list")
	comm.Dir = templateLocation

	stdoutStderr, err := comm.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}

	workspaceFound := false
	test := strings.Split(string(stdoutStderr), "\n")

	for _, t := range test {
		strippedT := strings.ReplaceAll(t, " ", "")
		strippedT = strings.ReplaceAll(strippedT, "*", "")
		if strippedT == projectName {
			workspaceFound = true
		}
	}

	if workspaceFound {
		comm = exec.Command("terraform", "workspace", "select", projectName)
		comm.Dir = templateLocation

		stdoutStderr, err := comm.CombinedOutput()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("%s\n", stdoutStderr)
	} else {
		comm = exec.Command("terraform", "workspace", "new", projectName)
		comm.Dir = templateLocation

		stdoutStderr, err := comm.CombinedOutput()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("%s\n", stdoutStderr)
	}

	return nil
}

func terraformApply(argSlice []interface{}, project map[string]interface{}) error {
	arguments := []string{}

	arguments = append(arguments, "apply")
	arguments = append(arguments, "-auto-approve")

	for _, arg := range argSlice {
		argMap := arg.(map[string]interface{})
		value := ""

		if project[argMap["variable"].(string)] != nil {
			value = project[argMap["variable"].(string)].(string)
		} else if argMap["default"] != nil {
			value = argMap["default"].(string)
		}
		a := fmt.Sprintf("-var=%s=%s", argMap["prefix"], value)

		arguments = append(arguments, a)
	}

	comm := exec.Command("terraform", arguments...)
	comm.Dir = templateLocation

	stdoutStderr, err := comm.CombinedOutput()
	if err != nil {
		fmt.Printf("%s\n", stdoutStderr)
		log.Fatal(err)
	}

	fmt.Printf("%s\n", stdoutStderr)

	return nil
}