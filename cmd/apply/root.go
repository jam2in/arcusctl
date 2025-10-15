package apply

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/jam2in/arcus-cli/internal"
	"github.com/jam2in/arcus-cli/internal/aclgroup"
	"github.com/spf13/cobra"
)

var (
	RootCmd = &cobra.Command{
		Use:   "apply FILENAME",
		Short: "TODO",
		Run:   runApply,
	}
)

func runApply(cmd *cobra.Command, args []string) {
	filename := args[0]

	f, err := os.Open(filename)
	if err != nil {
		fmt.Printf("failed to open %s: %v\n", filename, err)
		os.Exit(1)
	}
	defer f.Close()

	tasks, err := loadTasks(f)
	if err != nil {
		fmt.Printf("failed to load tasks: %v\n", err)
		os.Exit(1)
	}

	if len(tasks) == 0 {
		fmt.Println("no tasks to apply")
		return
	}

	for _, task := range tasks {
		fmt.Println(task.Description())
	}

	var yn string
	fmt.Printf("\napply? [y/N]: ")
	fmt.Scanln(&yn)
	if yn != "y" {
		fmt.Println("operation cancelled")
		os.Exit(1)
	}

	for _, task := range tasks {
		if err := task.Execute(); err != nil {
			fmt.Printf("%v\n", err)
		}
	}
}

func loadTasks(reader io.Reader) ([]internal.Task, error) {
	decoder := yaml.NewDecoder(reader)
	var tasks []internal.Task

	for {
		var node ast.Node
		if err := decoder.Decode(&node); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("decode: %w", err)
		}

		var k struct{ Kind string }
		if err := decoder.DecodeFromNode(node, &k); err != nil {
			return nil, fmt.Errorf("decode kind: %w", err)
		}

		switch k.Kind {
		case "aclgroup", "AclGroup":
			var r aclgroup.Resource
			if err := decoder.DecodeFromNode(node, &r); err != nil {
				return nil, fmt.Errorf("decode aclgroup: %w", err)
			}

		default:
			return nil, fmt.Errorf("unknown kind: %v", k.Kind)
		}
	}
}
