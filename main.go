package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/xanzy/go-gitlab"
	"log"
	"os"
	"path/filepath"
)

type Output struct {
	AllFiles             []string `json:"all_files"`
	AddedAndChangedFiles []string `json:"added_and_changed_files"`
	AddedFiles           []string `json:"added_files"`
	ChangedFiles         []string `json:"changed_files"`
	DeletedFiles         []string `json:"deleted_files"`
	RenamedFiles         []string `json:"renamed_files"`
	AnyAdded             bool     `json:"any_added"`
	AnyChanged           bool     `json:"any_changed"`
	AnyDeleted           bool     `json:"any_deleted"`
	AnyRenamed           bool     `json:"any_renamed"`
	OnlyAdded            bool     `json:"only_added"`
	OnlyChanged          bool     `json:"only_changed"`
	OnlyDeleted          bool     `json:"only_deleted"`
	OnlyRenamed          bool     `json:"only_renamed"`
	TypeChangedFiles     []string `json:"type_changed_files"`
}

func main() {
	var sourceBranch string
	var targetBranch string
	var isStraight bool
	var projectId int

	flag.StringVar(&sourceBranch, "source-branch", "", "Branch on which the changes exist.")
	flag.StringVar(&targetBranch, "target-branch", "", "Branch with which to compare.")
	flag.BoolVar(&isStraight, "straight", false, "Git comparison type, false for 'three dots comparison'.")
	flag.IntVar(&projectId, "project-id", 0, "Project in which the branches reside.")

	flag.Parse()
	var gitlabApiToken = os.Getenv("GITLAB_API_TOKEN")
	if gitlabApiToken == "" {
		log.Fatal("Variable 'GITLAB_API_TOKEN' must be set.")
	}

	git, err := gitlab.NewClient(gitlabApiToken)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	co := &gitlab.CompareOptions{
		From:     &sourceBranch,
		To:       &targetBranch,
		Straight: &isStraight,
	}
	compare, _, err := git.Repositories.Compare(projectId, co)
	if err != nil {
		log.Fatalf(
			"Failed to compare %v with %v. Ensure both branches exist. Error: %v", sourceBranch, targetBranch, err,
		)
	}
	Diffs := compare.Diffs

	var allFiles = make([]string, 0)
	var addedAndChangedFiles = make([]string, 0)
	var addedFiles = make([]string, 0)
	var changedFiles = make([]string, 0)
	var deletedFiles = make([]string, 0)
	var renamedFiles = make([]string, 0)
	var anyAdded = false
	var anyChanged = false
	var anyDeleted = false
	var anyRenamed = false
	var onlyAdded = true
	var onlyChanged = true
	var onlyDeleted = true
	var onlyRenamed = true
	var typeChangedFiles = make([]string, 0)

	for _, d := range Diffs {
		changedFile := !d.DeletedFile && !d.NewFile && !d.RenamedFile

		allFiles = append(allFiles, d.NewPath)

		if changedFile || d.NewFile {
			addedAndChangedFiles = append(addedAndChangedFiles, d.NewPath)
		}

		if d.NewFile {
			addedFiles = append(addedFiles, d.NewPath)
		}

		if changedFile {
			changedFiles = append(changedFiles, d.NewPath)
		}

		if d.DeletedFile {
			deletedFiles = append(deletedFiles, d.NewPath)
		}

		if d.RenamedFile {
			renamedFiles = append(renamedFiles, d.NewPath)
		}

		anyAdded = anyAdded || d.NewFile
		anyChanged = anyChanged || changedFile
		anyDeleted = anyDeleted || d.DeletedFile
		anyRenamed = anyRenamed || d.RenamedFile

		onlyAdded = onlyAdded && d.NewFile
		onlyChanged = onlyChanged && changedFile
		onlyDeleted = onlyDeleted && d.DeletedFile
		onlyRenamed = onlyRenamed && d.RenamedFile

		oldFileExtension := filepath.Ext(d.OldPath)
		newFileExtension := filepath.Ext(d.NewPath)
		if oldFileExtension != newFileExtension {
			typeChangedFiles = append(typeChangedFiles, d.NewPath)
		}
	}

	output := Output{
		AllFiles:             allFiles,
		AddedAndChangedFiles: addedAndChangedFiles,
		AddedFiles:           addedFiles,
		ChangedFiles:         changedFiles,
		DeletedFiles:         deletedFiles,
		RenamedFiles:         renamedFiles,
		AnyAdded:             anyAdded,
		AnyChanged:           anyChanged,
		AnyDeleted:           anyDeleted,
		AnyRenamed:           anyRenamed,
		OnlyAdded:            onlyAdded,
		OnlyChanged:          onlyChanged,
		OnlyDeleted:          onlyDeleted,
		OnlyRenamed:          onlyRenamed,
		TypeChangedFiles:     typeChangedFiles,
	}

	content, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		panic(err)
	} else {
		fmt.Println(string(content))
	}
}
